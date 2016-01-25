package queue

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/viper"
	"strings"
)

type SQSQueue struct{}

func (q *SQSQueue) Get() (*ScriptCommand, error) {
	sess := session.New(&aws.Config{
		Region: aws.String(viper.GetString("queue.region")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("credentials.aws_access_id"),
			viper.GetString("credentials.aws_secret"),
			"",
		),
	})

	svc := sqs.New(sess)

	params := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(viper.GetString("queue.url")),
		AttributeNames:        []*string{},
		MaxNumberOfMessages:   aws.Int64(1),
		MessageAttributeNames: []*string{},
		VisibilityTimeout:     aws.Int64(1),
	}

	resp, err := svc.ReceiveMessage(params)
	if len(resp.Messages) > 0 {
		dec := json.NewDecoder(strings.NewReader(*resp.Messages[0].Body))
		var cmd *ScriptCommand
		err = dec.Decode(&cmd)
		cmd.Receipt = *resp.Messages[0].ReceiptHandle
		return cmd, err
	} else {
		return nil, nil
	}
}
func (q *SQSQueue) Put(mailbox string, cmd *ScriptCommand) error {
	sess := session.New(&aws.Config{
		Region: aws.String(viper.GetString("queue.region")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("credentials.aws_access_id"),
			viper.GetString("credentials.aws_secret"),
			"",
		),
	})

	svc := sqs.New(sess)
	params := &sqs.SendMessageInput{
		QueueUrl:          aws.String(viper.GetString("queue.url")),
		MessageBody:       aws.String(cmd.ToJson()),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{},
	}
	_, err := svc.SendMessage(params)
	return err
}

func (q *SQSQueue) Delete(cmd *ScriptCommand) error {
	sess := session.New(&aws.Config{
		Region: aws.String(viper.GetString("queue.region")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("credentials.aws_access_id"),
			viper.GetString("credentials.aws_secret"),
			"",
		),
	})

	svc := sqs.New(sess)
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(viper.GetString("queue.url")),
		ReceiptHandle: &cmd.Receipt,
	}
	_, err := svc.DeleteMessage(params)
	return err
}
