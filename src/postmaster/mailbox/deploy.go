package mailbox

import (
	"github.com/cznic/ql"
	"github.com/nu7hatch/gouuid"
	"time"
)

func GetOpenDeployments() ([]Deployment, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT id, name, deployedAt, deployedBy, totalMessages
		FROM deployment
		WHERE open == true
	`)
	if err != nil {
		return nil, err
	}
	deployments := []Deployment{}
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		deployment := Deployment{
			Id:            data[0].(string),
			Name:          data[1].(string),
			DeployedAt:    data[2].(time.Time),
			DeployedBy:    data[3].(string),
			TotalMessages: data[4].(int64),
		}
		deployments = append(deployments, deployment)
		return true, nil
	})
	return deployments, nil
}

func FindDeployment(id string) (*Deployment, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT id, name, deployedAt, deployedBy, open, messageBody
		FROM deployment
		WHERE id == $1;
		SELECT count(*) FROM message WHERE deployment == $1;
	`, id)
	if err != nil {
		return nil, err
	}
	var deployment *Deployment
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		deployment = &Deployment{
			Id:          data[0].(string),
			Name:        data[1].(string),
			DeployedAt:  data[2].(time.Time),
			DeployedBy:  data[3].(string),
			Open:        data[4].(bool),
			MessageBody: data[5].(string),
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

func CreateDeployment(name string, apiToken string,
	message string) (*Deployment, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	deploymentName := name
	if name == "" {
		deploymentName = id.String()
	}
	dp := &Deployment{
		Id:         id.String(),
		Name:       deploymentName,
		DeployedBy: apiToken,
		DeployedAt: time.Now(),
		Open:       true,
	}
	_, _, err = DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO deployment (
			id, name, deployedAt, deployedBy, totalMessages, open, messageBody
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		);
		COMMIT;
	`, dp.Id, dp.Name, dp.DeployedAt, dp.DeployedBy, dp.TotalMessages, dp.Open,
		message)
	if err != nil {
		return nil, err
	}
	return dp, nil
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

func (dp *Deployment) GetName() string {
	if dp.Name == "" {
		return dp.Id
	} else {
		return dp.Name
	}
}

func (dp *Deployment) GetResponses() ([]DeploymentResponse, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT deployment, mailbox, response, respondedAt
		FROM deploymentResponse
		WHERE deployment == $1`, dp.Id)
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
		}
		responses = append(responses, response)
		return true, nil
	})
	return responses, nil
}

func (dp *Deployment) Stats() (*DeploymentStats, error) {
	resp, _, err := DB.Run(ql.NewRWCtx(), `
		SELECT count(*)
		FROM message
		WHERE deployment == $1 AND deleted == false;
		SELECT totalMessages
		FROM deployment
		WHERE id == $1;
		SELECT count(*)
		FROM deploymentResponse
		WHERE deployment == $1;`, dp.Id)
	if err != nil {
		return nil, err
	}
	stats := &DeploymentStats{}
	resp[0].Do(false, func(data []interface{}) (bool, error) {
		stats.PendingCount = data[0].(int64)
		return false, nil
	})
	resp[1].Do(false, func(data []interface{}) (bool, error) {
		stats.MessageCount = data[0].(int64)
		return false, nil
	})
	resp[2].Do(false, func(data []interface{}) (bool, error) {
		stats.ResponseCount = data[0].(int64)
		return false, nil
	})
	return stats, nil
}

func (dp *Deployment) AddResponse(mailbox, response string) error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
		BEGIN TRANSACTION;
		INSERT INTO deploymentResponse (
			deployment, mailbox, response, respondedAt
		) VALUES (
			$1,$2,$3,$4
		);
		COMMIT;
	`, dp.Id, mailbox, response, time.Now())
	return err
}

type DeploymentResponse struct {
	Deployment  string
	Mailbox     string
	Response    string
	RespondedAt time.Time
}

type DeploymentStats struct {
	MessageCount  int64
	PendingCount  int64
	ResponseCount int64
}
