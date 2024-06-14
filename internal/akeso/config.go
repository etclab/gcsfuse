package akeso

import (
	"fmt"
)

type Config struct {
	Strategy  string
	AkesoDir  string
	ProjectID string
	TopicID   string
	SubID     string
	Key       []byte
}

func (c *Config) String() string {
	return fmt.Sprintf("AkesoConfig{Strategy: %q, AkesoDir: %q, ProjectId: %q, TopicID: %q, SubID: %q}",
		c.Strategy, c.AkesoDir, c.ProjectID, c.TopicID, c.SubID)
}
