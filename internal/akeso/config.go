package akeso

import (
    "fmt"
)

type Config struct {
    AkesoDir    string
    Strategy    string
    PubSub      string

    Key         []byte
}

func (c *Config) String() string {
    return fmt.Sprintf("AkesoConfig{AkesoDir: %q, Strategy: %q, PubSub: %q}", c.AkesoDir, c.Strategy, c.PubSub)
}
