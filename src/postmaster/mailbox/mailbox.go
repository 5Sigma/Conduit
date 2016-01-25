package mailbox

import (
	"github.com/cznic/ql"
	"github.com/nu7hatch/gouuid"
	"time"
)

type Mailbox struct {
	Id string
}

func (mb *Mailbox) PutMessage(body string) (*Message, error) {
	id, _ := uuid.NewV4()
	msg := &Message{
		Id:        id.String(),
		Body:      body,
		MailboxId: mb.Id,
		CreatedAt: time.Now(),
	}
	db, err := OpenDB()
	if err != nil {
		return nil, err
	}
	_, _, err = db.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			INSERT INTO message (
				id, receiveCount, body, mailbox
			) VALUES (
				$1, $2, $3, $4
			);
			COMMIT;
		`, msg.Id, msg.ReceiveCount, msg.Body, msg.MailboxId)
	db.Close()
	return msg, err
}

func (mb *Mailbox) GetMessage() (*Message, error) {
	db, err := OpenDB()
	if db != nil {
		defer db.Close()
	}
	if err != nil {
		return nil, err
	}

	rss, _, err := db.Run(ql.NewRWCtx(), `
		SELECT id, receiveCount, body, mailbox, createdAt
		FROM message
		WHERE mailbox == $1
		LIMIT 1;
		`, mb.Id)
	var msg Message
	if len(rss) > 0 {
		r, _ := rss[0].FirstRow()
		if r == nil {
			return nil, nil
		}
		rss[0].Do(false, func(data []interface{}) (bool, error) {
			msg = Message{
				Id:           data[0].(string),
				ReceiveCount: data[1].(int64),
				Body:         data[2].(string),
				MailboxId:    data[3].(string),
				CreatedAt:    data[4].(time.Time),
			}
			return false, nil
		})
	} else {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	msg.ReceiveCount++

	_, _, err = db.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		UPDATE message
		SET receiveCount = $1
		WHERE id == $2;
		COMMIT;
	`, msg.ReceiveCount, msg.Id)
	if err != nil {
		return nil, err
	}
	return &msg, err
}

func (mb *Mailbox) MessageCount() (int64, error) {
	db, err := OpenDB()
	if db != nil {
		defer db.Close()
	}
	rss, _, err := db.Run(ql.NewRWCtx(), `
		SELECT count()
		FROM message
		WHERE mailbox == $1
	`, mb.Id)
	if err != nil {
		return -1, err
	}
	var count int64
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		count = data[0].(int64)
		return false, nil
	})
	return count, nil
}

func Create() (*Mailbox, error) {
	id, _ := uuid.NewV4()
	mb := &Mailbox{Id: id.String()}
	db, err := OpenDB()
	if db != nil {
		defer db.Close()
	}
	if err != nil {
		return nil, err
	}
	_, _, err = db.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO mailbox (
			id
		) VALUES (
			$1
		);
		COMMIT
		`, mb.Id)
	if err != nil {
		return nil, err
	}
	return mb, nil
}

func All() ([]Mailbox, error) {
	var mbxs []Mailbox
	db, err := OpenDB()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	rss, _, err := db.Run(ql.NewRWCtx(), `
		SELECT id FROM mailbox`)
	if err != nil {
		return nil, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		mb := Mailbox{
			Id: data[0].(string),
		}
		mbxs = append(mbxs, mb)
		return true, nil
	})
	return mbxs, nil
}

func Find(id string) (*Mailbox, error) {
	var mbx *Mailbox
	db, err := OpenDB()
	if db != nil {
		defer db.Close()
	}
	if err != nil {
		return nil, err
	}
	rss, _, err := db.Run(ql.NewRWCtx(),
		` SELECT id FROM mailbox WHERE id==$1`, id)
	if err != nil {
		return nil, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		mbx = &Mailbox{
			Id: data[0].(string),
		}
		return false, nil
	})
	return mbx, nil
}
