package mailbox

import (
	"errors"
	"github.com/cznic/ql"
)

type AccessToken struct {
	Token      string
	FullAccess bool
	Mailbox    string
	Name       string
}

func (token *AccessToken) Create() error {
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

func (mb *Mailbox) CreateToken() (*AccessToken, error) {
	token := &AccessToken{
		Token:      GenerateIdentifier(),
		Mailbox:    mb.Id,
		FullAccess: false,
	}
	err := token.Create()
	return token, err
}

func CreateAPIToken(name string) (*AccessToken, error) {
	id := GenerateIdentifier()
	tokenName := name
	if name == "" {
		tokenName = id
	}
	token := &AccessToken{
		Token:      id,
		FullAccess: true,
		Name:       tokenName,
	}
	err := token.Create()
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
