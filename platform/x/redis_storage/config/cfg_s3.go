package config

import (
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

type S3OnlandConfig struct {
	AWS_Region    string `mapstructure:"AWS_Region,omitempty"`
	AWS_AccessKey string `mapstructure:"AWS_AccessKey,omitempty"`
	AWS_SecretKey string `mapstructure:"AWS_SecretKey,omitempty"`

	S3_Bucket string `mapstructure:"Bucket"`
	Format    string `mapstructure:"Onland_Format,omitempty"`
	FormatSep string `mapstructure:"Onland_Format_Separator,omitempty"`
}

func (c *S3OnlandConfig) HasConfigured() bool {
	if c.S3_Bucket == "" ||
		len(c.S3_Bucket) == 0 {
		return false
	}
	return true
}

func init() {
	Register("S3", &S3OnlandConfig{})
}

func (c *S3OnlandConfig) Setup(addFn func(storehelper.IStore)) {
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

	s3_store := storehelper.NewStore(storehelper.S3, storehelper.S3Cfg{
		Region:           aws_region,
		BucketsVerUpdate: c.S3_Bucket,
		AccessKeyId:      aws_accessKey,
		SecretAccessKey:  aws_accessKey,
		Format:           format,
		Seq:              formatSep,
	})
	addFn(s3_store)

}
