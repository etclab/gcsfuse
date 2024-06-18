package akeso

import (
	"fmt"
	"sync"
)

type Config struct {
	Strategy  string
	AkesoDir  string
	ProjectID string
	TopicID   string
	SubID     string
	Key       []byte
	mu        sync.Mutex
}

func (c *Config) String() string {
	return fmt.Sprintf("AkesoConfig{Strategy: %q, AkesoDir: %q, ProjectId: %q, TopicID: %q, SubID: %q}",
		c.Strategy, c.AkesoDir, c.ProjectID, c.TopicID, c.SubID)
}

func (c *Config) SetKey(key []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Key = key
}
