package queue

import (
	"encoding/json"
	"fmt"
	"postmaster/api"
	"postmaster/client"
)

type Queue struct {
	Client client.Client
}

func New(host string, mailbox string, token string) Queue {
	return Queue{
		Client: client.Client{
			Host:    host,
			Mailbox: mailbox,
			Token:   token,
		},
	}
}

func (cmd *ScriptCommand) ToJson() string {
	bytes, _ := json.Marshal(cmd)
	return string(bytes)
}

func (q *Queue) getMailboxUrl() string {
	return fmt.Sprintf("http://%s/get", q.Client.Host)
}

func (q *Queue) Get() (*ScriptCommand, error) {

	resp, err := q.Client.Get()
	if err != nil {
		return nil, err
	}
	if resp.IsEmpty() {
		return nil, nil
	}
	script := &ScriptCommand{
		ScriptBody: resp.Body,
		Receipt:    resp.Message,
		Deployment: resp.Deployment,
	}
	return script, nil
}

func (q *Queue) Put(mailboxes []string, pattern string,
	cmd *ScriptCommand, deployment string) (int, error) {
	resp, err := q.Client.Put(mailboxes, pattern, cmd.ScriptBody)
	if err != nil {
		return 0, err
	}
	return len(resp.Mailboxes), err
}

func (q *Queue) Delete(cmd *ScriptCommand) error {
	_, err := q.Client.Delete(cmd.Receipt)
	return err
}

func (q *Queue) SystemStats() (*api.SystemStatsResponse, error) {
	return q.Client.Stats()
}

type ScriptCommand struct {
	ScriptBody string
	Receipt    string
	Deployment string
}
