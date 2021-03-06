package mailbox

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cznic/ql"
	"github.com/nu7hatch/gouuid"
	"strings"
	"time"
)

type (

	// Mailboxes represent a bucket or queue of messages. Messages can be added to
	// the mailbox through a deployment. Messages can be requested from the mailbox
	// and one will be returned (with no garenteed ordering). Once the message has
	// been processed it can be deleted from the mailbox.
	//
	// Mailboxes must have a unique Id, but this value can be anything unique. The
	// system is designed with the idea of manually created and namespaced
	// identifiers such as:
	//
	//		newton.maxwell.bohr
	//
	// This allows pattern searches to be intuitive such as:
	//
	//		newton.*.bohr
	Mailbox struct {
		Id       string
		LastSeen time.Time
		Version  string
		Host     string
	}

	SystemStats struct {
		MailboxCount    int64
		MessageCount    int64
		PendingMessages int64
	}
)

// GenerateIdentifier is used for generating various IDs. It is used to create
// messageIds, deploymentIds, access tokens, etc.
func GenerateIdentifier() string {
	uuid, _ := uuid.NewV4()
	return hex.EncodeToString(uuid[0:16])
}

// All returns a slice of all mailboxes.
func All() ([]Mailbox, error) {
	var mbxs []Mailbox
	rss, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT id, lastCheckedInAt, version, host FROM mailbox`)
	if err != nil {
		return nil, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		mb := Mailbox{
			Id: data[0].(string),
		}
		if data[1] != nil {
			mb.LastSeen = data[1].(time.Time)
		}
		if data[2] != nil {
			mb.Version = data[2].(string)
		}
		if data[3] != nil {
			mb.Host = data[3].(string)
		}
		mbxs = append(mbxs, mb)
		return true, nil
	})
	return mbxs, nil
}

// Find will return a mailbox or nil for a given mailbox identifier
func Find(id string) (*Mailbox, error) {
	var mbx *Mailbox
	rss, _, err := DB.Run(ql.NewRWCtx(),
		` SELECT id, lastCheckedInAt FROM mailbox WHERE id==$1`, id)
	if err != nil {
		return nil, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		mbx = &Mailbox{
			Id: data[0].(string),
		}
		if data[1] != nil {
			mbx.LastSeen = data[1].(time.Time)
		}
		return false, nil
	})
	return mbx, nil
}

// Marks a message as deleted in the mailbox.
func DeleteMessage(msgId string) error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
	BEGIN TRANSACTION;
	UPDATE message
	SET deleted = true
	WHERE id == $1;
	COMMIT;
	`, msgId)
	return err
}

// Create will generate and persist a new mailbox.
func Create(id string) (*Mailbox, error) {
	var mb *Mailbox
	mb, err := Find(id)
	if err != nil {
		return nil, err
	}
	if mb != nil {
		return mb, errors.New("Mailbox already exists")
	}
	mb = &Mailbox{
		Id: strings.ToLower(id),
	}
	_, _, err = DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT 	INTO mailbox (id)
		VALUES 	($1);
		COMMIT;
		`, mb.Id)
	if err != nil {
		return nil, err
	}
	return mb, nil
}

func Deregister(id string) error {
	mb, err := Find(id)
	if err != nil {
		return err
	}
	if mb == nil {
		return errors.New("Mailbox not found")
	}
	_, _, err = DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		DELETE FROM mailbox
		WHERE id == $1;
		DELETE FROM message
		WHERE mailbox == $1;
		DELETE FROM accessToken
		WHERE mailbox ==  $1;
		COMMIT;
		`, mb.Id)
	return err
}

// FindMessage will return a Message or nil for a given message identifier
func FindMessage(msgId string) (*Message, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT  message.id, deployment.messageBody, message.mailbox,
		message.deployment, message.receiveCount, message.lastReceivedAt,
		message.createdAt, message.deleted
		FROM message, deployment
		WHERE 
			message.id == $1
			AND message.deleted == false
			AND deployment.id == message.deployment
		LIMIT 1;
	`, msgId)
	if err != nil {
		return nil, err
	}
	var message *Message
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		message = readMessageStruct(data)
		return false, nil
	})
	return message, nil
}

// Search will return a slice of mailboxes whos identifiers match a given
// pattern. This pattern can be any valid regex. However it will automatically
// convert '*' to '.*'. This allows * to be used as a simple wildcard when
// searching by pattern.
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
		return true, nil
	})
	return mbxs, nil
}

// PutMessage will automatically generate a deployment and add a message to the
// mailbox. The access token for the deployment will be "SYSTEM". This is
// intended to be used for internal actions.
func (mb *Mailbox) PutMessage(body string) (*Message, error) {
	dep := &Deployment{
		MessageBody: body,
		DeployedBy:  "SYSTEM",
	}
	err := dep.Create()
	if err != nil {
		return nil, err
	}
	msg, err := mb.DeployMessage(dep.Id)
	return msg, err
}

// DeployMessage accepts a deployment identifier and adds this mailbox to its
// deployment. A new message will be available for this deployment in the
// mailbox.
func (mb *Mailbox) DeployMessage(depId string) (*Message, error) {
	deployment, err := FindDeployment(depId)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, errors.New(fmt.Sprintf("Deployment %s not found", depId))
	}

	msg := &Message{
		Id:         GenerateIdentifier(),
		Mailbox:    mb.Id,
		CreatedAt:  time.Now(),
		Deployment: deployment.Id,
		Body:       deployment.MessageBody,
	}

	_, _, err = DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			INSERT INTO message (
				id, receiveCount, mailbox, createdAt, deployment, deleted
			) VALUES (
				$1, $2, $3, $4, $5, false
			);
			UPDATE deployment
			SET totalMessages = totalMessages + 1
			WHERE id == $5;
			COMMIT;
		`, msg.Id, msg.ReceiveCount, msg.Mailbox, msg.CreatedAt,
		deployment.Id)
	return msg, err
}

func (mb *Mailbox) Checkin(host, version string) error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		UPDATE 	mailbox
		SET 		lastCheckedInAt = $1, host = $2, version = $3
		WHERE 	id == $4;
		COMMIT;
		`, time.Now(), host, version, mb.Id)
	return err
}

// GetMessage returns a message from the mailbox. Once the message is processed
// it should be removed from the queue with Delete.
func (mb *Mailbox) GetMessage() (*Message, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT  message.id, deployment.messageBody, message.mailbox,
		message.deployment, message.receiveCount, message.lastReceivedAt,
		message.createdAt, message.deleted
		FROM message, deployment
		WHERE 
			message.mailbox == $1
			AND message.deleted == false
			AND deployment.id == message.deployment
		LIMIT 1;
		`, mb.Id)
	if err != nil {
		return nil, err
	}
	var msg *Message
	if len(rss) > 0 {
		r, _ := rss[0].FirstRow()
		if r == nil {
			return nil, nil
		}
		rss[0].Do(false, func(data []interface{}) (bool, error) {
			msg = readMessageStruct(data)
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
	return msg, err
}

// Purge will mark all messages in the mailbox as deleted. They will no longer
// be availble when polling for messages.
func (mb *Mailbox) Purge() (int64, error) {
	ctx := ql.NewRWCtx()
	_, _, err := DB.Run(ctx, `
	BEGIN TRANSACTION;
	UPDATE message
	SET deleted = true
	WHERE mailbox == $1;
	COMMIT;
	`, mb.Id)
	return ctx.RowsAffected, err
}

// MessageCount returns a cound of all pending messages in the mailbox. This
// will not return messages that were marked as deleted.
func (mb *Mailbox) MessageCount() (int64, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT count()
		FROM message
		WHERE mailbox == $1 AND deleted == false
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

// Stats returns a SystemStats structure with overall message count information.
func Stats() (*SystemStats, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT count(*) FROM message where deleted == false;
		SELECT count(*) FROM message;
		SELECT count(*) FROM mailbox;`)
	if err != nil {
		return nil, err
	}
	stats := &SystemStats{}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		stats.PendingMessages = data[0].(int64)
		return false, nil
	})
	rss[1].Do(false, func(data []interface{}) (bool, error) {
		stats.MessageCount = data[0].(int64)
		return false, nil
	})
	rss[2].Do(false, func(data []interface{}) (bool, error) {
		stats.MailboxCount = data[0].(int64)
		return false, nil
	})
	return stats, nil
}

func AssetPending(md5 string) (bool, error) {
	rss, _, err := DB.Run(ql.NewRWCtx(), `
	SELECT 		count(message.id)
	FROM 			message, deployment
	WHERE 		message.deployment == deployment.id
	AND 			message.deleted == false
	AND 			deployment.asset == $1`, md5)
	if err != nil {
		return false, err
	}
	var count int64 = 1
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		count = data[0].(int64)
		return true, nil
	})
	return count != 0, nil
}
