package mailbox

import (
	"fmt"
	"github.com/cznic/ql"
	"github.com/kardianos/osext"
	"os"
	"path/filepath"
)

var DB *ql.DB

func CreateDB() error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			CREATE TABLE message (
				id string,
				receiveCount int,
				mailbox string,
				createdAt time,
				lastReceivedAt time,
				deployment string,
				deleted bool
			);
			CREATE TABLE mailbox (
				id string,
				completedMessages int,
				createdAt time,
				lastCompletedAt time,
				lastCheckedInAt time
			);
			CREATE TABLE accessToken (
				mailbox string,
				token string,
				name string,
				fullAccess bool
			);
			CREATE TABLE deployment (
				id string,
				messageBody string,
				name string,
				deployedAt time,
				deployedBy string,
				totalMessages int,
				open bool
			);
			CREATE TABLE deploymentResponse (
				deployment string,
				mailbox string,
				response string,
				respondedAt time
			);
			INSERT INTO accessToken (mailbox, token, name, fullAccess)
			VALUES ("conduit.system", "SYSTEM", "conduit.system", false);
			COMMIT;`)
	return err
}

func OpenDB() {
	var err error
	shouldCreate := false
	directory, _ := osext.ExecutableFolder()
	path := filepath.Join(directory, "mailboxes.db")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		shouldCreate = true
	}
	DB, err = ql.OpenFile(path, &ql.Options{CanCreate: true})
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
