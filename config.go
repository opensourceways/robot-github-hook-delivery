package main

import (
	"github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	Config config.MQConfig `json:"config" required:"true"`
}

func (c *configuration) Validate() error {
	return nil
}

func (c *configuration) SetDefault() {
}
