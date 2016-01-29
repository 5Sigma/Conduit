package mailbox

import (
	"github.com/cznic/ql"
	"github.com/nu7hatch/gouuid"
	"strings"
	"time"
)

type Mailbox struct {
	Id string
}

func (mb *Mailbox) PutMessage(body string) (*Message, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	msg := &Message{
		Id:        id.String(),
		Body:      body,
		MailboxId: mb.Id,
		CreatedAt: time.Now(),
	}
	_, _, err = DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			INSERT INTO message (
				id, receiveCount, body, mailbox, createdAt
			) VALUES (
				$1, $2, $3, $4, $5
			);
			COMMIT;
		`, msg.Id, msg.ReceiveCount, msg.Body, msg.MailboxId, time.Now())
	return msg, err
}

func (mb *Mailbox) GetMessage() (*Message, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
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

	_, _, err = DB.Run(ql.NewRWCtx(), `
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

func DeleteMessage(msgId string) error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
	BEGIN TRANSACTION;
	DELETE FROM message
	WHERE id == $1;
	COMMIT;
	`, msgId)
	return err
}

func (mb *Mailbox) Purge() (int64, error) {
	ctx := ql.NewRWCtx()
	_, _, err := DB.Run(ctx, `
	BEGIN TRANSACTION;
	DELETE FROM message
	WHERE mailbox == $1;
	COMMIT;
	`, mb.Id)
	return ctx.RowsAffected, err
}

func (mb *Mailbox) MessageCount() (int64, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
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

func Create(id string) (*Mailbox, error) {
	mb := &Mailbox{Id: strings.ToLower(id)}
	_, _, err := DB.Run(ql.NewRWCtx(), `
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

func (mb *Mailbox) Stats() (*MailboxStats, error) {
	count, err := mb.MessageCount()
	if err != nil {
		return nil, err
	}
	stats := &MailboxStats{
		MailboxId:       mb.Id,
		PendingMessages: count,
	}
	return stats, nil
}

func All() ([]Mailbox, error) {
	var mbxs []Mailbox
	rss, _, err := DB.Run(ql.NewRWCtx(), `
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
	rss, _, err := DB.Run(ql.NewRWCtx(),
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

func Search(rawPattern string) ([]Mailbox, error) {
	mbxs := []Mailbox{}
	pattern := strings.ToLower(strings.Replace(rawPattern, "*", ".*", -1))
	rss, _, err := DB.Run(ql.NewRWCtx(),
		` SELECT id FROM mailbox WHERE id LIKE $1`, pattern)
	if err != nil {
		return nil, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		mb := Mailbox{Id: data[0].(string)}
		mbxs = append(mbxs, mb)
		return false, nil
	})
	return mbxs, nil
}

type MailboxStats struct {
	MailboxId       string
	PendingMessages int64
}
