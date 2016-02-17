package mailbox

import (
	"github.com/cznic/ql"
	"time"
)

type Message struct {
	Id             string
	Body           string
	ReceiveCount   int64
	Mailbox        string
	CreatedAt      time.Time
	LastReceivedAt time.Time
	Deployment     string
	Deleted        bool
}

func (m *Message) Create() error {
	if m.Id == "" {
		m.Id = GenerateIdentifier()
	}
	_, _, err := DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			INSERT INTO message (
				id, receiveCount, mailbox, createdAt, deployment, deleted
			) VALUES (
				$1, $2, $3, $4, $5, false
			);
			COMMIT;
		`, m.Id, m.ReceiveCount, m.Mailbox, m.CreatedAt, m.Deployment)
	return err
}
func (m *Message) Save() error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		UPDATE message
		SET receiveCount = $2, mailbox = $3, createdAt = $4, deployment = $5,
			deleted = $6
		WHERE id = $1`, m.Id, m.Mailbox, m.CreatedAt, m.Deployment, m.Deleted)
	return err
}

func (m *Message) GetDeployment() (*Deployment, error) {
	return FindDeployment(m.Deployment)
}

// readMessageStruct is used to read the row data into a Message
func readMessageStruct(data []interface{}) *Message {
	message := &Message{
		Id:           data[0].(string),
		Body:         data[1].(string),
		Mailbox:      data[2].(string),
		Deployment:   data[3].(string),
		ReceiveCount: data[4].(int64),
		Deleted:      data[7].(bool),
		CreatedAt:    data[6].(time.Time),
	}
	if data[5] != nil {
		message.LastReceivedAt = data[5].(time.Time)
	}
	return message
}
