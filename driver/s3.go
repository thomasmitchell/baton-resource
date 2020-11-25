package driver

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/thomasmitchell/baton-resource/models"
)

type s3Driver struct {
	s3     *s3.S3
	bucket string
}

func newS3Driver(cfg models.Source) (*s3Driver, error) {
	session := session.Must(session.NewSession())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TLSSkipVerify,
			},
		},
	}

	awsCfg := aws.NewConfig().
		WithCredentials(
			credentials.NewStaticCredentials(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		).
		WithEndpoint(cfg.Endpoint).
		WithRegion(cfg.Region).
		WithHTTPClient(client)

	return &s3Driver{
			s3:     s3.New(session, awsCfg),
			bucket: cfg.Bucket,
		},
		nil
}

func (s *s3Driver) Read(key string) (*models.Payload, error) {
	getObjOut, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		ioutil.ReadAll(getObjOut.Body)
		getObjOut.Body.Close()
	}()

	dec := json.NewDecoder(getObjOut.Body)
	ret := models.Payload{}
	err = dec.Decode(&ret)
	return &ret, err
}

func (s *s3Driver) Write(key string, payload models.Payload) error {
	jBuf, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	_, err = s.s3.PutObject(&s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   bytes.NewReader(jBuf),
	})

	return err
}

func (s *s3Driver) ReadVersion(key string) (*models.Version, error) {
	getObjOut, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return &models.Version{}, nil
			}
		}
		return nil, err
	}
	defer func() {
		ioutil.ReadAll(getObjOut.Body)
		getObjOut.Body.Close()
	}()

	dec := json.NewDecoder(getObjOut.Body)
	ret := models.Payload{}
	err = dec.Decode(&ret)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response as JSON: %s", err)
	}

	return &ret.Version, nil
}
