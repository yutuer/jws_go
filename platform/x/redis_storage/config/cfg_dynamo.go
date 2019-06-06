package config

import "vcs.taiyouxi.net/platform/planx/util/storehelper"

type DynamoDBOnlandConfig struct {
	AWS_Region    string `mapstructure:"AWS_Region,omitempty"`
	AWS_AccessKey string `mapstructure:"AWS_AccessKey,omitempty"`
	AWS_SecretKey string `mapstructure:"AWS_SecretKey,omitempty"`

	DynamoDB  string `mapstructure:"DB"`
	Format    string `mapstructure:"Onland_Format"`
	FormatSep string `mapstructure:"Onland_Format_Separator"`
}

func (c *DynamoDBOnlandConfig) HasConfigured() bool {
	if c.DynamoDB == "" ||
		len(c.DynamoDB) == 0 {
		return false
	}
	return true
}

func init() {
	Register("DynamoDB", &DynamoDBOnlandConfig{})
}

func (c *DynamoDBOnlandConfig) Setup(addFn func(storehelper.IStore)) {
	if !c.HasConfigured() {
		return
	}

	aws_region := c.AWS_Region
	aws_accessKey := c.AWS_AccessKey
	aws_secretKey := c.AWS_SecretKey
	format := c.Format
	formatSep := c.FormatSep

	if aws_region == "" {
		aws_region = CommonCfg.AWS_Region
	}

	if aws_accessKey == "" {
		aws_accessKey = CommonCfg.AWS_Region
	}

	if aws_secretKey == "" {
		aws_secretKey = CommonCfg.AWS_Region
	}

	if format == "" {
		format = CommonCfg.Format
	}

	if formatSep == "" {
		formatSep = CommonCfg.FormatSep
	}

	dynamodb_store := storehelper.NewStoreDynamoDB(
		aws_region,
		c.DynamoDB,
		aws_accessKey,
		aws_secretKey,
		format,
		formatSep)
	addFn(dynamodb_store)

}
