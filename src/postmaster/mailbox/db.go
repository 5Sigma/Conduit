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

func OpenDB() {
	var err error
	shouldCreate := false
	if _, err := os.Stat("mailboxes.db"); os.IsNotExist(err) {
		shouldCreate = true
	}
	DB, err = ql.OpenFile("mailboxes.db", &ql.Options{CanCreate: true})
	if err != nil {
		fmt.Println("Could not open mailbox database.")
		os.Exit(-1)
	}
	if shouldCreate {
		err := CreateDB()
		if err != nil {
			panic(err)
		}
	}
}

func OpenMemDB() error {
	var err error
	DB, err = ql.OpenMem()
	return err
}

func CloseDB() error {
	err := DB.Close()
	return err
}
