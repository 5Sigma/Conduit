package mailbox

import (
	"github.com/cznic/ql"
	"time"
)

type Message struct {
	Id             string
	Body           string
	ReceiveCount   int64
	MailboxId      string
	CreatedAt      time.Time
	lastReceivedAt time.Time
}

func (msg *Message) Completed() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	_, _, err = db.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		DELETE FROM message
		WHERE id = $2;
		COMMIT
		`, msg.ReceiveCount+1)
	return err

}
