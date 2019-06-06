package awshelper

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateAWSSession(region, accessKey, secretKey string, MaxRetries int) *session.Session {
	config := defaults.Config().WithRegion(region).WithMaxRetries(MaxRetries)
	handler := defaults.Handlers()

	providers := []credentials.Provider{
		&credentials.StaticProvider{Value: credentials.Value{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			SessionToken:    "",
		}},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
		defaults.RemoteCredProvider(*config, handler),
	}

	if accessKey == "" ||
		secretKey == "" {
		providers = providers[1:]
	}

	creds := credentials.NewChainCredentials(providers)
	//TODO: For FanYang by YZH， WithDisableComputeChecksums还是必须的吗？
	config = config.
		WithCredentials(creds).
		WithDisableSSL(true).
		WithDisableComputeChecksums(true)

	return session.New(config)

}
