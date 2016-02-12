package mailbox

import (
	"fmt"
	"github.com/cznic/ql"
	"time"
)

type (
	DeploymentResponse struct {
		Deployment  string
		Mailbox     string
		Response    string
		RespondedAt time.Time
		IsError     bool
	}

	DeploymentStats struct {
		MessageCount  int64
		PendingCount  int64
		ResponseCount int64
	}
)

func ListDeployments(name string, count int, token string) ([]Deployment, error) {
	sql := fmt.Sprintf(`
		SELECT 		deployment.id, deployment.name, deployment.deployedAt,
							accessToken.name, deployment.totalMessages
		FROM  		deployment, accessToken
		WHERE 		deployment.name LIKE $1
			AND 		deployment.deployedBy LIKE $2
			AND 		accessToken.name == deployment.deployedBy
		ORDER BY 	deployment.deployedAt DESC
		LIMIT %d
		`, count)
	resp, _, err := DB.Run(ql.NewRWCtx(), sql, name, token)
	if err != nil {
		return nil, err
	}
	deployments := []Deployment{}
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		deployment := Deployment{
			Id:            data[0].(string),
			Name:          data[1].(string),
			DeployedBy:    data[3].(string),
			TotalMessages: data[4].(int64),
		}
		if data[2] != nil {
			deployment.DeployedAt = data[2].(time.Time)
		}
		deployments = append(deployments, deployment)
		return true, nil
	})
	return deployments, nil
}

func FindDeployment(id string) (*Deployment, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	id, name, deployedAt, deployedBy, open, messageBody
		FROM 		deployment
		WHERE 	id == $1;

		SELECT 	count(*) 
		FROM 		message
		WHERE 	deployment == $1;
	`, id)
	if err != nil {
		return nil, err
	}
	var deployment *Deployment
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		deployment = &Deployment{
			Id:          data[0].(string),
			Name:        data[1].(string),
			DeployedBy:  data[3].(string),
			Open:        data[4].(bool),
			MessageBody: data[5].(string),
		}
		if data[2] != nil {
			deployment.DeployedAt = data[2].(time.Time)
		}
		return false, nil
	})
	if deployment != nil {
		resp[1].Do(false, func(data []interface{}) (bool, error) {
			deployment.TotalMessages = data[0].(int64)
			return false, nil
		})
	}
	return deployment, nil
}

type Deployment struct {
	Id            string
	Name          string
	DeployedAt    time.Time
	DeployedBy    string
	TotalMessages int64
	MessageBody   string
	Open          bool
}

func (dp *Deployment) Save() error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		UPDATE 	deployment
		SET 		name = $2, deployedAt = $3, deployedBy = $4, totalMessages = $5,
						messageBody = $6
		WHERE 	id == $1;
		COMMIT;
	`, dp.Id, dp.Name, dp.DeployedAt, dp.DeployedBy, dp.TotalMessages, dp.MessageBody)
	return err
}

func (dp *Deployment) Create() error {
	dp.Id = GenerateIdentifier()
	if dp.Name == "" {
		dp.Name = dp.Id
	}

	dp.DeployedAt = time.Now()

	if dp.DeployedBy == "" {
		dp.DeployedBy = "SYSTEM"
	}

	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO 	deployment 
									(id, name, deployedBy, totalMessages, messageBody,
									open, deployedAt)
		VALUES 				($1, $2, $3, 0, $4, true, $5);
		COMMIT;
	`, dp.Id, dp.Name, dp.DeployedBy, dp.MessageBody, dp.DeployedAt)
	return err
}

func (dp *Deployment) GetName() string {
	if dp.Name == "" {
		return dp.Id
	} else {
		return dp.Name
	}
}

func (dp *Deployment) Deploy(mb *Mailbox) (*Message, error) {
	msg := &Message{
		Id:         GenerateIdentifier(),
		Mailbox:    mb.Id,
		CreatedAt:  time.Now(),
		Deployment: dp.Id,
		Body:       dp.MessageBody,
	}
	err := msg.Create()
	dp.TotalMessages++
	err = dp.Save()
	return msg, err
}

func (dp *Deployment) GetResponses() ([]DeploymentResponse, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	deployment, mailbox, response, respondedAt, isError
		FROM 		deploymentResponse
		WHERE 	deployment == $1
		`, dp.Id)
	if err != nil {
		return nil, err
	}
	responses := []DeploymentResponse{}
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		response := DeploymentResponse{
			Deployment:  data[0].(string),
			Mailbox:     data[1].(string),
			Response:    data[2].(string),
			RespondedAt: data[3].(time.Time),
			IsError:     data[4].(bool),
		}
		responses = append(responses, response)
		return true, nil
	})
	return responses, nil
}

func (dp *Deployment) Stats() (*DeploymentStats, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT 	count(*)
		FROM 		message
		WHERE 	deployment == $1 AND deleted == false;

		SELECT 	count(*)
		FROM 		deploymentResponse
		WHERE 	deployment == $1;`, dp.Id)
	if err != nil {
		return nil, err
	}
	stats := &DeploymentStats{}
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		stats.PendingCount = data[0].(int64)
		return false, nil
	})
	resp[1].Do(false, func(data []interface{}) (bool, error) {
		stats.ResponseCount = data[0].(int64)
		return false, nil
	})
	stats.MessageCount = dp.TotalMessages
	return stats, nil
}

func (dp *Deployment) AddResponse(mailbox, response string, isErr bool) error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO 	deploymentResponse 
									(deployment, mailbox, response, respondedAt, isError)
		VALUES 				($1,$2,$3,$4,$5);
		COMMIT;
	`, dp.Id, mailbox, response, time.Now(), isErr)
	return err
}
