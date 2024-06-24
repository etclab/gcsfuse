package akeso

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/etclab/art"
	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

func CreateDirsIfNotExist(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("os.MkdirAll(%q) failed: %w", dir, err)
		}
	}
	return nil
}

func SavePubIKFile(ikKeyPair []byte, fileName string) error {
	var keyPair KeyPairMessage
	err := json.Unmarshal(ikKeyPair, &keyPair)
	if err != nil {
		return fmt.Errorf("error unmarshalling public IK: %v", err)
	}

	pubKey := ed25519.PublicKey(keyPair.PublicKey)
	err = art.WritePublicIKToFile(pubKey, fileName, art.EncodingPEM)
	if err != nil {
		return fmt.Errorf("error writing public IK file: %v", err)
	}

	return nil
}

func SavePrivEKFile(ekKeyPair []byte, fileName string) error {
	var keyPair KeyPairMessage
	err := json.Unmarshal(ekKeyPair, &keyPair)
	if err != nil {
		return fmt.Errorf("error parsing private EK: %v", err)
	}

	privKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("error decoding private key: %v", err)
	}

	privKey, err := ecdh.X25519().NewPrivateKey(privKeyBytes)
	if err != nil {
		return fmt.Errorf("error saving private key: %v", err)
	}
	err = art.WritePrivateEKToFile(privKey, fileName, art.EncodingPEM)
	if err != nil {
		return fmt.Errorf("error writing public IK file: %v", err)
	}

	return nil
}

func SaveSetupMsg(setupMsg art.SetupMessage, fileName string) error {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "    ")
	enc.Encode(setupMsg)

	setupMsgBytes := buf.Bytes()

	err := os.WriteFile(fileName, setupMsgBytes, 0666)
	if err != nil {
		return fmt.Errorf("error writing setup message file: %v", err)
	}

	return nil
}

func SavePubsubMessage(msg *pubsub.Message, dir string) {
	fileName := msg.PublishTime.Format(time.RFC822) + "-" + msg.ID + ".json"
	filePath := filepath.Join(dir, fileName)

	var bytes bytes.Buffer
	enc := json.NewEncoder(&bytes)
	enc.SetIndent("", "    ")
	enc.Encode(msg)

	err := os.WriteFile(filePath, bytes.Bytes(), 0666)
	if err != nil {
		logger.Errorf("error saving pubsub message file: %v", err)
	}
}

func PublishMessage(ctx context.Context, data []byte, attrs map[string]string,
	config *Config) {
	projectID := config.ProjectID
	topicID := config.TopicID

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		logger.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	topic := client.Topic(topicID)

	res := topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attrs,
	})

	id, err := res.Get(ctx)
	if err != nil {
		logger.Errorf("Failed to publish: %v", err)
	}

	logger.Infof("Published message with msg ID: %v\n", id)
}

func RemoveFileIfExists(fileName string) {
	err := os.Remove(fileName)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// ignore if the file doesn't exist
		logger.Infof("error removing file: %v", err)
	}
	// todo: handle other errors
}
