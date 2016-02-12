package agent

import (
	"conduit/agent"
	"conduit/engine"
	"postmaster/mailbox"
	"testing"
)

func TestAgentCommand(t *testing.T) {
	mailbox.OpenMemDB()
	key := mailbox.AccessKey{FullAccess: true}
	key.Create()
	agent.Address = ":6112"
	agent.AccessKey = key.Secret
	go agent.Start()
	engine.Agents["test"] = ":6112"
	engine.AgentAccessKey = key.Secret
	eng := engine.New()
	err := eng.Execute(`$agent("test", function() { console.log("test"); });`)
	if err != nil {
		t.Fatal(err)
	}
}
