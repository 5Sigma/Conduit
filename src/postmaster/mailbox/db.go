package mailbox

import (
	"github.com/cznic/ql"
)

func CreateDB() error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	_, _, err = db.Run(ql.NewRWCtx(), `
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
				id string
			);
			COMMIT;`)
	db.Close()
	return err
}

func OpenDB() (*ql.DB, error) {
	return ql.OpenFile("mailboxes.db", &ql.Options{CanCreate: true})
}
