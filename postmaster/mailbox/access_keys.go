package mailbox

import (
	"errors"
	"github.com/cznic/ql"
)

type AccessKey struct {
	Name       string
	MailboxId  string
	FullAccess bool
	Secret     string
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
	if KeyExists(key.Name) {
		return errors.New("Key already exists")
	}
	key.Secret = GenerateIdentifier()
	_, _, err := DB.Run(ql.NewRWCtx(), `
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
		SELECT name, mailbox, fullAccess, token
		FROM accessToken
		WHERE name == $1
		LIMIT 1;
	`, name)
	if err != nil {
		return nil, err
	}
	key := &AccessKey{}
	res[0].Do(false, func(data []interface{}) (bool, error) {
		ql.Unmarshal(key, data)
		return false, nil
	})
	return key, nil
}

func KeyExists(name string) bool {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	count(*)
		FROM 		accessToken
		WHERE 	name == $1
		LIMIT 	1;
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

func AllKeys() ([]*AccessKey, error) {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	name, mailbox, fullAccess, token
		FROM 		accessToken`)
	if err != nil {
		return nil, err
	}
	keys := []*AccessKey{}
	err = res[0].Do(false, func(data []interface{}) (bool, error) {
		key := &AccessKey{}
		if err := ql.Unmarshal(key, data); err != nil {
			return false, err
		}
		keys = append(keys, key)
		return true, nil
	})
	return keys, nil
}

func AdminKeys() ([]*AccessKey, error) {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	name, mailbox, fullAccess, token
		FROM 		accessToken
		WHERE 	fullAccess == true
	`)
	if err != nil {
		return nil, err
	}
	keys := []*AccessKey{}
	err = res[0].Do(false, func(data []interface{}) (bool, error) {
		key := &AccessKey{}
		if err := ql.Unmarshal(key, data); err != nil {
			return false, err
		}
		keys = append(keys, key)
		return true, nil
	})
	return keys, nil
}

func Revoke(name string) error {
	ctx := ql.NewRWCtx()
	_, _, err := DB.Run(ctx, `
		BEGIN TRANSACTION;
		DELETE FROM 		accessToken
		WHERE 					name == $1;
		COMMIT;
	`, name)
	if err != nil {
		return err
	}
	if ctx.RowsAffected == 0 {
		return errors.New("Key not found")
	}
	return nil
}
