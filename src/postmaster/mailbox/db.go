package mailbox

import (
	"fmt"
	"github.com/cznic/ql"
	"os"
)

var DB *ql.DB

func CreateDB() error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			CREATE TABLE message (
				id string,
				receiveCount int,
				body string,
				mailbox string,
				createdAt time,
				lastReceivedAt time
			);
			CREATE TABLE mailbox (
				id string,
				completedMessages int,
				createdAt time,
				lastCompletedAt time,
				lastCheckedInAt time
			);
			CREATE TABLE tokens (
				mailboxId string,
				token string,
				secret string,
				canSend bool
			);
			COMMIT;`)
	return err
}

func init() {
	var err error
	DB, err = ql.OpenFile("mailboxes.db", &ql.Options{CanCreate: true})
	if err != nil {
		fmt.Println("Could not open mailbox database.")
		os.Exit(-1)
	}
}

func CloseDB() error {
	err := DB.Close()
	return err
}
