package mailbox

import (
	"errors"
	"github.com/cznic/ql"
	"github.com/nu7hatch/gouuid"
)

type AccessToken struct {
	Token      string
	FullAccess bool
	Mailbox    string
	Name       string
}

func saveToken(token *AccessToken) error {
	if token.FullAccess == false {
		mb, err := Find(token.Mailbox)
		if err != nil {
			return err
		}
		if mb == nil {
			return errors.New("Mailbox not found.")
		}
	}
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO accessToken (
			mailbox, token, fullAccess, name
		) VALUES (
			$1, $2, $3, $4
		);
		COMMIT;
	`, token.Mailbox, token.Token, token.FullAccess, token.Name)
	return err
}

func CreateMailboxToken(mailbox string) (*AccessToken, error) {
	id, _ := uuid.NewV4()
	token := &AccessToken{
		Token:      id.String(),
		Mailbox:    mailbox,
		FullAccess: false,
	}
	err := saveToken(token)
	return token, err
}

func CreateAPIToken(name string) (*AccessToken, error) {
	id, _ := uuid.NewV4()
	if name == "" {
		nameUUID, _ := uuid.NewV4()
		name = nameUUID.String()
	}
	token := &AccessToken{
		Token:      id.String(),
		FullAccess: true,
		Name:       name,
	}
	err := saveToken(token)
	return token, err
}

func findToken(tokenStr string) (*AccessToken, error) {
	res, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT name, token, mailbox, fullAccess
		FROM accessToken
		WHERE token == $1
	`, tokenStr)
	if err != nil {
		return nil, err
	}
	var token *AccessToken
	res[0].Do(false, func(data []interface{}) (bool, error) {
		token = &AccessToken{
			Name:       data[0].(string),
			Token:      data[1].(string),
			Mailbox:    data[2].(string),
			FullAccess: data[3].(bool),
		}
		return false, nil
	})
	return token, nil
}

func TokenCanPut(tokenStr string, mailbox string) bool {
	token, err := findToken(tokenStr)
	if err != nil || token == nil {
		return false
	}
	return token.FullAccess
}

func TokenCanGet(tokenStr string, mailbox string) bool {
	token, err := findToken(tokenStr)
	if err != nil || token == nil {
		return false
	}
	if token.FullAccess {
		return true
	}
	if token.Mailbox == mailbox {
		return true
	}
	return false
}

func TokenCanAdmin(tokenStr string) bool {
	token, err := findToken(tokenStr)
	if err != nil || token == nil {
		return false
	}
	return token.FullAccess
}

func TokenCanRegister(tokenStr string) bool {
	token, err := findToken(tokenStr)
	if err != nil || token == nil {
		return false
	}
	return token.FullAccess
}
