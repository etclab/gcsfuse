package akeso

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/googlecloudplatform/gcsfuse/v2/internal/logger"
)

func subscriptionPullLoop(ctx context.Context, sub *pubsub.Subscription,
	config *Config) {
	for {
		err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			logger.Infof("received pubsub message (%d bytes): %s", len(msg.Data), string(msg.Data))
			key, err := hex.DecodeString(string(msg.Data))
			if err != nil {
				logger.Warnf("hex.DecodeString() failed: %v", err)
			}
			config.SetKey(key)
			msg.Ack()
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
