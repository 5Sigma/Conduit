package mailbox

import (
	"postmaster/mailbox"
	"testing"
)

func TestAddResponse(t *testing.T) {
	mb, err := mailbox.Create("tests.AddResponses")
	if err != nil {
		t.Fatal(err)
	}
	token, err := mailbox.CreateAPIToken("AddResponseToken")
	if err != nil {
		t.Fatal(err)
	}
	dep, err := mailbox.CreateDeployment("test", token.Token, "test")
	if err != nil {
		t.Fatal(err)
	}
	msg, err := mb.DeployMessage(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = dep.AddResponse(msg.Id, "response")
	if err != nil {
		t.Fatal(err)
	}
	responses, err := dep.GetResponses()
	if len(responses) == 0 {
		t.Fatal("Deployment has no responses")
	}
}
