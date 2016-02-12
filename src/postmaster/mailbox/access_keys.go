package mailbox

import (
	"errors"
	"github.com/cznic/ql"
)

type AccessKey struct {
	Secret     string
	FullAccess bool
	MailboxId  string
	Name       string
}

func (key *AccessKey) Create() error {
	if key.FullAccess == false {
		mb, err := Find(key.MailboxId)
		if err != nil {
			return err
		}
		if mb == nil {
			return errors.New("Can't generate key. Mailbox not found.")
		}
		key.Name = key.MailboxId
	} else {
		if key.Name == "" {
			key.Name = GenerateIdentifier()
		}
	}
	k, err := FindKeyByName(key.Name)
	if err != nil {
		return err
	}
	if k != nil {
		return errors.New("Key already exists")
	}
	key.Secret = GenerateIdentifier()
	_, _, err = DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO accessToken (
			mailbox, token, fullAccess, name
		) VALUES (
			$1, $2, $3, $4
		);
		COMMIT;
	`, key.MailboxId, key.Secret, key.FullAccess, key.Name)
	return err
}

func (key *AccessKey) CanPut(mb *Mailbox) bool {
	return key.FullAccess
}

func (key *AccessKey) CanAdmin() bool {
	return key.FullAccess
}

func (key *AccessKey) CanDelete(mb *Mailbox) bool {
	return key.CanGet(mb)
}

func (key *AccessKey) CanGet(mb *Mailbox) bool {
	if key.FullAccess {
		return true
	}
	if key.MailboxId == mb.Id {
		return true
	}
	return false
}

func FindKeyByName(name string) (*AccessKey, error) {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT name, token, mailbox, fullAccess
		FROM accessToken
		WHERE name == $1
		LIMIT 1;
	`, name)
	if err != nil {
		return nil, err
	}
	var key *AccessKey
	res[0].Do(false, func(data []interface{}) (bool, error) {
		key = &AccessKey{
			Name:       data[0].(string),
			Secret:     data[1].(string),
			MailboxId:  data[2].(string),
			FullAccess: data[3].(bool),
		}
		return false, nil
	})
	return key, nil
}

func KeyExists(name string) bool {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT count(*)
		FROM accessToken
		WHERE name == $1
		LIMIT 1;
	`, name)
	if err != nil {
		return false
	}
	var count int64
	res[0].Do(false, func(data []interface{}) (bool, error) {
		count = data[0].(int64)
		return false, nil
	})
	return count == 1
}
