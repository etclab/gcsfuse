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
	KeyMutex  sync.RWMutex
}

func (c *Config) String() string {
	return fmt.Sprintf("AkesoConfig{Strategy: %q, AkesoDir: %q, ProjectId: %q, TopicID: %q, SubID: %q}",
		c.Strategy, c.AkesoDir, c.ProjectID, c.TopicID, c.SubID)
}

func (c *Config) SetKey(key []byte) {
	c.KeyMutex.Lock()
	defer c.KeyMutex.Unlock()
	c.Key = key
}
