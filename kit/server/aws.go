package server

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
)

// InitAWSSession initializes an AWS session.
func InitAWSSession(cfg *CommonConfig) *session.Session {
	return session.New(&aws.Config{HTTPClient: MakeHTTPClientForConfig(cfg)})
}

// InitAWSSES initializes a SES client.
func InitAWSSES(cfg *CommonConfig) *ses.SES {
	return ses.New(InitAWSSession(cfg))
}

// InitAWSSNS initializes a SNS client.
func InitAWSSNS(cfg *CommonConfig) *sns.SNS {
	return sns.New(InitAWSSession(cfg))
}

// InitAWSKinesis initializes a Kinesis client.
func InitAWSKinesis(cfg *CommonConfig) *kinesis.Kinesis {
	return kinesis.New(InitAWSSession(cfg))
}
