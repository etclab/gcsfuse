package akeso

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/etclab/art"
	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

func waitForKeyUpdateMessages(ctx context.Context, sub *pubsub.Subscription,
	config *Config) {
	for {
		err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			logger.Infof("received pubsub message id %s", msg.ID)
			SavePubsubMessage(msg, config.ArtConfig.pubsubMsgDir)

			// maybe use only one pubsub topic?
			// check if the tree state exists - if not we can't update the key
			// this handles the case when member is starting up for the first time
			// and the update msg is received before the setup msg
			stateFile := config.ArtConfig.treeStateFile
			if !FileExists(stateFile) {
				logger.Errorf("tree state file not found: %v", stateFile)
				return
			}

			attrs := msg.Attributes
			msgType, ok := attrs["messageType"]

			if ok {
				if msgType == "update_key" {
					// TODO: this will be triggered by a timer or some other event
					// TODO: for now we send a dummy message of type update_key to trigger this

					msgFor, ok := attrs["messageFor"]
					if ok && msgFor == config.ArtConfig.memberName {
						// member updates their key and broadcasts the update message
						msgData := updateKey(msg, config)
						msgAttrs := map[string]string{
							"messageType": "update_msg",
							"updatedBy":   config.ArtConfig.memberName,
						}

						PublishMessage(ctx, msgData, msgAttrs, config)
					} else {
						logger.Infof("skipping update_key message meant for: %v", msgFor)
					}
				} else if msgType == "update_msg" {
					if attrs["updatedBy"] == config.ArtConfig.memberName {
						logger.Infof("skipping own update message")
					} else {
						processUpdateMessage(msg, config)
					}
				} else {
					logger.Errorf("Unknown message type: %v", msgType)
				}

				msg.Ack()
			} else {
				logger.Errorf("Missing message type for topic: %v", config.TopicID)
			}

		})
		if err != nil {
			logger.Warnf("sub.Receive() failed: %v .. shutting down", err)
			os.Exit(1) // TODO: is there a more graceful way?
		}
	}
}

func subscriptionPullLoop(ctx context.Context, sub *pubsub.Subscription,
	config *Config) {
	for {
		err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			logger.Infof("received pubsub message id %s", msg.ID)
			SavePubsubMessage(msg, config.ArtConfig.pubsubMsgDir)

			err := processSetupMessage(msg, config)
			if err != nil {
				logger.Errorf("processSetupMessage failed: %v", err)
			} else {
				msg.Ack()
			}
		})
		if err != nil {
			logger.Warnf("sub.Receive() failed: %v .. shutting down", err)
			os.Exit(1) // TODO: is there a more graceful way?
		}
	}
}

func StartSubscriptionPullLoop(config *Config) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient failed: %w", err)
	}

	sub := client.Subscription(config.SubID)
	go subscriptionPullLoop(ctx, sub, config)

	return nil
}

// todo: refactor this with the above function
func StartKeyUpdateSubscriptionPullLoop(config *Config) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient failed: %w", err)
	}

	sub := client.Subscription(config.ArtConfig.keyUpdateSubID)
	go waitForKeyUpdateMessages(ctx, sub, config)

	return nil
}

// assuming no two members update their keys at the same time
func updateKey(msg *pubsub.Message, config *Config) []byte {
	logger.Infof("updating key on receiving message with id %s", msg.ID)

	ac := config.ArtConfig

	idx := ac.index
	stateFile := ac.treeStateFile
	stageKeyFile := ac.stageKeyFile

	updateMsgFile := ac.updateMsgFile
	updateMsgMacFile := ac.updateMsgMacFile

	updateMsg, state, stageKey := art.UpdateKey(idx, stateFile)

	updateMsg.Save(updateMsgFile)

	RemoveFileIfExists(updateMsgMacFile)
	updateMsg.SaveMac(*stageKey, updateMsgMacFile)

	// to atomically replace old state and stage key files
	// use overwrite-by-rename: https://unix.stackexchange.com/a/35289/460185
	state.Save(stateFile)

	RemoveFileIfExists(stageKeyFile)
	state.SaveStageKey(stageKeyFile)

	// update the stage key
	stageKeyBytes, err := os.ReadFile(stageKeyFile)
	if err != nil {
		logger.Errorf("os.ReadFile(stageKeyFile) failed: %v", err)
	}
	config.SetKey(stageKeyBytes)

	macBytes, err := os.ReadFile(updateMsgMacFile)
	if err != nil {
		logger.Errorf("os.ReadFile(updateMsgMacFile) failed: %v", err)
	}

	keyUpdateMsg := &KeyUpdateMessage{
		UpdateMsg:    *updateMsg,
		UpdateMsgMac: macBytes,
		UpdatedBy:    ac.memberName,
	}

	msgBytes, err := json.Marshal(keyUpdateMsg)
	if err != nil {
		logger.Errorf("json.Marshal(keyUpdateMsg) failed: %v", err)
	}

	return msgBytes
}

func processSetupMessage(msg *pubsub.Message, config *Config) error {
	// TODO: load config from yml file
	ac := config.ArtConfig

	idx := ac.index
	memberName := ac.memberName

	initiatorPubIKFile := ac.initiatorPubIKFile
	privEKFile := ac.memberPrivEKFile

	setupMsgFile := ac.setupMsgFile
	sigFile := ac.setupMsgSigFile

	stateFile := ac.treeStateFile
	stageKeyFile := ac.stageKeyFile

	var data GroupSetupMessage
	err := json.Unmarshal(msg.Data, &data)

	if err != nil {
		logger.Errorf("error unmarshalling message: %v", err)
	}

	setupMsg := data.SetupMsg
	signature := data.SetupMsgSig

	err = SaveSetupMsg(setupMsg, setupMsgFile)
	if err != nil {
		logger.Errorf("SaveSetupMsg failed: %v", err)
	}

	err = os.WriteFile(sigFile, signature, 0666)
	if err != nil {
		logger.Errorf("error writing sign file: %v", err)
	}

	// save initiator public keys to file
	err = SavePubIKFile(data.InPubKey, initiatorPubIKFile)
	if err != nil {
		logger.Errorf("SavePubIKFile failed: %v", err)
	}

	// now for the private key
	memberEK, ok := data.EKeys[memberName]
	if !ok {
		logger.Errorf("no private EK found for member %v", memberName)
	}

	err = SavePrivEKFile(memberEK, privEKFile)
	if err != nil {
		logger.Errorf("SavePrivEKFile failed: %v", err)
	}

	state := art.ProcessSetupMessage(idx, privEKFile, setupMsgFile,
		initiatorPubIKFile, sigFile)

	state.Save(stateFile)
	state.SaveStageKey(stageKeyFile)

	stageKey, err := os.ReadFile(stageKeyFile)
	if err != nil {
		logger.Errorf("os.ReadFile(stageKeyFile) failed: %v", err)
	}

	config.SetKey(stageKey)

	return nil
}

func processUpdateMessage(msg *pubsub.Message, config *Config) error {
	logger.Infof("processing update message with id %s", msg.ID)

	data := msg.Data

	var keyUpdateMsg KeyUpdateMessage
	err := json.Unmarshal(data, &keyUpdateMsg)
	if err != nil {
		logger.Errorf("json.Unmarshal(data, &keyUpdateMsg) failed: %v", err)
	}

	// todo: does it affect the stage and stage key by reapplying its own update?
	// todo: if applying the update here skip the state and stage key update on update_key
	if keyUpdateMsg.UpdatedBy == config.ArtConfig.memberName {
		logger.Infof("skipping own update message")
		return nil
	}

	ac := config.ArtConfig

	idx := ac.index
	stateFile := ac.treeStateFile
	stageKeyFile := ac.stageKeyFile

	updateMsgFile := ac.updateMsgFile
	updateMsgMacFile := ac.updateMsgMacFile

	err = SaveUpdateMsg(keyUpdateMsg.UpdateMsg, updateMsgFile)
	if err != nil {
		logger.Errorf("SaveUpdateMsg failed: %v", err)
	}

	RemoveFileIfExists(updateMsgMacFile)
	err = os.WriteFile(updateMsgMacFile, keyUpdateMsg.UpdateMsgMac, 0666)
	if err != nil {
		logger.Errorf("error writing mac file: %v", err)
	}

	state := art.ProcessUpdateMessage(idx, stateFile, updateMsgFile, updateMsgMacFile)

	state.Save(stateFile)
	RemoveFileIfExists(stageKeyFile)
	state.SaveStageKey(stageKeyFile)

	return nil
}
