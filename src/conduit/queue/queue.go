package queue

import (
	"encoding/json"
	"github.com/spf13/viper"
)

func GetQueue() Queue {
	if viper.GetString("queue.service") == "sqs" {
		return &SQSQueue{}
	}

	if viper.GetString("queue.service") == "conduit" {
		return &ConduitQueue{}
	}

	return nil
}

type ScriptCommand struct {
	RemoteScriptUrl string   `json: "url"`
	RemoteAssets    []string `json: "assets"`
	ScriptBody      []byte   `json: "body"`
	Receipt         string   `json: "receipt"`
}

func (cmd *ScriptCommand) ToJson() string {
	bytes, _ := json.Marshal(cmd)
	return string(bytes)
}

type Queue interface {
	Get() (*ScriptCommand, error)
	Put(mailbox string, cmd *ScriptCommand) error
	Delete(cmd *ScriptCommand) error
}
