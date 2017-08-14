package mailbox

import (
	"testing"
)

func TestAddResponse(t *testing.T) {
	mb, err := Create("tests.AddResponses")
	if err != nil {
		t.Fatal(err)
	}
	key := AccessKey{FullAccess: true}
	key.Create()
	if err != nil {
		t.Fatal(err)
	}
	dep := Deployment{MessageBody: "test", DeployedBy: key.Name}
	err = dep.Create()
	if err != nil {
		t.Fatal(err)
	}
	msg, err := mb.DeployMessage(dep.Id)
	if err != nil {
		t.Fatal(err)
	}
	err = dep.AddResponse(msg.Id, "response", false)
	if err != nil {
		t.Fatal(err)
	}
	responses, err := dep.GetResponses()
	if len(responses) == 0 {
		t.Fatal("Deployment has no responses")
	}
}
