package agent

import (
	"github.com/5sigma/conduit/engine"
	"github.com/5sigma/conduit/postmaster/mailbox"
	"testing"
)

func TestAgentCommand(t *testing.T) {
	mailbox.OpenMemDB()
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	Address = ":6112"
	AccessKey = key.Secret
	go Start()
	engine.Agents["test"] = ":6112"
	engine.AgentAccessKey = key.Secret
	eng := engine.New()
	err := eng.Execute(`$agent("test", function() { console.log("test"); });`)
	if err != nil {
		t.Fatal(err)
	}
}
