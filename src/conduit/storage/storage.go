package storage

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"io/ioutil"
)

type StorageManager interface {
	GetScript(filename string) (string, error)
	GetAsset(filename string) (string, error)
	PutScript(filename string) error
	PutAsset(filename string) error
}

type S3StorageManager struct {
}

func (mgr *S3StorageManager) getSession() *session.Session {
	return session.New(&aws.Config{
		Region: aws.String(viper.GetString("queue.region")),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("credentials.aws_access_id"),
			viper.GetString("credentials.aws_secret"),
			"",
		),
	})
}

func (mgr *S3StorageManager) getService() *s3.S3 {
	sess := mgr.getSession()
	return s3.New(sess)
}

func (mgr *S3StorageManager) uploadFile(filename string, key string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	bucket := viper.GetString("storage.bucket")
	params := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	svc := mgr.getService()
	_, err = svc.PutObject(params)
	return err
}

func (mgr *S3StorageManager) downloadFile(key, string, path string) error {
	bucket := viper.GetString("storage.bucket")
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	svc := mgr.getService()
	_, err := svc.GetObject(params)
	return err
}

func (mgr *S3StorageManager) PutScript(filename string) error {
	prefix := viper.GetString("storage.script_prefix")
	key := prefix + filename
	err := mgr.uploadFile(filename, key)
	return err
}

func (mgr *S3StorageManager) PutAsset(filename string) error {
	prefix := viper.GetString("storage.asset_prefix")
	key := prefix + filename
	mgr.uploadFile(key, filename)
	return nil
}

func (mgr *S3StorageManager) GetScript(name string) (string, error) {
	prefix := viper.GetString("storage.script_prefix")
	key := prefix + name
	bucket := viper.GetString("storage.bucket")
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	svc := mgr.getService()
	resp, err := svc.GetObject(params)
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err

}

func (mgr *S3StorageManager) GetAsset(name string) (string, error) {
	prefix := viper.GetString("storage.script_prefix")
	key := prefix + name
	bucket := viper.GetString("storage.bucket")
	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	svc := mgr.getService()
	_, err := svc.GetObject(params)
	return "", err
}

func GetStorage() StorageManager {
	service := viper.GetString("storage.service")
	if service == "s3" {
		return &S3StorageManager{}
	}
	return nil
}
