package queue

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
)

type ConduitQueue struct{}

func getMailboxUrl() string {
	return fmt.Sprintf("%s/mailbox/%s", viper.GetString("queue.host"),
		viper.GetString("mailbox"))
}

func (q *ConduitQueue) Get() (*ScriptCommand, error) {
	resp, err := http.Get(getMailboxUrl())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	var cmd *ScriptCommand
	dec.Decode(&cmd)
	return cmd, nil
}

func (q *ConduitQueue) Put(mailbox string, cmd *ScriptCommand) error {
	var data []byte
	json.NewEncoder(bytes.NewBuffer(data)).Encode(cmd)
	req, err := http.NewRequest("POST", getMailboxUrl(), bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	} else {
		return errors.New("API Error: " + resp.Status)
	}
}

func (q *ConduitQueue) Delete(cmd *ScriptCommand) error {
	return nil
}
