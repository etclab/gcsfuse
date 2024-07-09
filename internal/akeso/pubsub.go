package akeso

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/etclab/art"
	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

func waitForKeyUpdateMessages(ctx context.Context, sub *pubsub.Subscription,
	config *Config) {
	ac := config.ArtConfig
	for {
		err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			// time starts here
			then := time.Now()

			logger.Infof("received pubsub message id %s", msg.ID)

			// maybe use only one pubsub topic?
			// check if the tree state exists - if not we can't update the key
			// this handles the case when member is starting up for the first time
			// and the update msg is received before the setup msg
			stateFile := ac.TreeStateFile
			if !FileExists(stateFile) {
				logger.Errorf("tree state file not found: %v", stateFile)

				msg.Nack()
				return
			}

			attrs := msg.Attributes
			msgType, ok := attrs["messageType"]

			// TODO: think about message ordering and what difference it makes
			if ok {
				if msgType == "update_key" {
					// TODO: this will be triggered by a timer or some other event
					// TODO: for now we send a dummy message of type update_key to trigger this

					msgFor, ok := attrs["messageFor"]
					if ok && msgFor == ac.MemberName {
						// member updates their key and broadcasts the update message
						msgData := updateKey(msg, config)

						// time taken to update its own key
						now := time.Now()
						diff := now.Sub(then)
						tag := fmt.Sprintf("update_key_%v", attrs["tag"])
						go saveTimeTaken(diff.Nanoseconds(), config, tag)

						msgAttrs := map[string]string{
							"messageType": "update_msg",
							"updatedBy":   ac.MemberName,
							"tag":         attrs["tag"],
							"OrderingKey": "update_messages",
						}

						PublishMessage(ctx, msgData, msgAttrs, config)
					} else {
						logger.Infof("skipping update_key message meant for: %v", msgFor)
					}
				} else if msgType == "update_msg" {
					if messageAlreadyProcessed(msg, config.PubSubDir) {
						logger.Infof("message id: %s has already been processed", msg.ID)

						msg.Ack()
						return
					}

					if attrs["updatedBy"] == ac.MemberName {
						logger.Infof("skipping own update message")
					} else {
						processUpdateMessage(msg, config)

						// time taken to apply the update_msg
						now := time.Now()
						diff := now.Sub(then)
						tag := fmt.Sprintf("update_msg_%v", attrs["tag"])
						go saveTimeTaken(diff.Nanoseconds(), config, tag)
					}
				} else {
					logger.Errorf("Unknown message type: %v", msgType)
				}

				SavePubsubMessage(msg, config.PubSubDir)
				msg.Ack()
			} else {
				logger.Errorf("Missing message type for topic: %v", config.UpdateTopicID)
			}

		})
		if err != nil {
			logger.Warnf("sub.Receive() failed: %v .. shutting down", err)
			os.Exit(1) // TODO: is there a more graceful way?
		}
	}
}

func saveTimeTaken(timeTaken int64, config *Config, tag string) {
	ac := config.ArtConfig
	fileName := filepath.Join(config.AkesoDir, ac.GroupName, ac.MemberName, tag)
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("os.OpenFile failed: %v", err)
	}

	_, err = f.WriteString(fmt.Sprintf("%v\n", timeTaken))
	if err != nil {
		logger.Errorf("f.WriteString failed: %v", err)
	}

	defer f.Close()
}

func subscriptionPullLoop(ctx context.Context, sub *pubsub.Subscription,
	config *Config) {
	// for {
	// TODO: ignore setup message if already there
	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		logger.Infof("received pubsub message id %s", msg.ID)
		SavePubsubMessage(msg, config.PubSubDir)

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
	// }
}

func StartSubscriptionPullLoop(config *Config) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient failed: %w", err)
	}

	sub := client.Subscription(config.SetupSubID)
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

	sub := client.Subscription(config.UpdateSubID)
	go waitForKeyUpdateMessages(ctx, sub, config)

	return nil
}

// // assuming no two members update their keys at the same time
func updateKey(msg *pubsub.Message, config *Config) []byte {
	logger.Infof("updating key on receiving message with id %s", msg.ID)

	ac := config.ArtConfig

	idx := ac.Index
	stateFile := ac.TreeStateFile
	stageKeyFile := ac.StageKeyFile

	updateMsgFile := ac.UpdateMsgFile
	updateMsgMacFile := ac.UpdateMsgMacFile

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
	config.SetKeyFile(stageKeyFile)

	macBytes, err := os.ReadFile(updateMsgMacFile)
	if err != nil {
		logger.Errorf("os.ReadFile(updateMsgMacFile) failed: %v", err)
	}

	keyUpdateMsg := &KeyUpdateMessage{
		UpdateMsg:    *updateMsg,
		UpdateMsgMac: macBytes,
		UpdatedBy:    ac.MemberName,
	}

	msgBytes, err := json.Marshal(keyUpdateMsg)
	if err != nil {
		logger.Errorf("json.Marshal(keyUpdateMsg) failed: %v", err)
	}

	return msgBytes
}

func processSetupMessage(msg *pubsub.Message, config *Config) error {
	ac := config.ArtConfig

	idx := ac.Index
	memberName := ac.MemberName

	initiatorPubIKFile := ac.InitiatorPubIKFile
	privEKFile := ac.MemberPrivEKFile

	setupMsgFile := ac.SetupMsgFile
	sigFile := ac.SetupMsgSigFile

	stateFile := ac.TreeStateFile
	stageKeyFile := ac.StageKeyFile

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

	config.SetKeyFile(stageKeyFile)

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
	if keyUpdateMsg.UpdatedBy == config.ArtConfig.MemberName {
		logger.Infof("skipping own update message")
		return nil
	}

	ac := config.ArtConfig

	idx := ac.Index
	stateFile := ac.TreeStateFile
	stageKeyFile := ac.StageKeyFile

	updateMsgFile := ac.UpdateMsgFile
	updateMsgMacFile := ac.UpdateMsgMacFile

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
