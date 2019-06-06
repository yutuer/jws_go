package gate

import "vcs.taiyouxi.net/platform/planx/util"

func AwsGetPublicIP() (string, error) {
	return util.AwsGetPublicIP()
}

func externalIP() (string, error) {
	return util.ExternalIP()
}

func GetPublicIP(pip, listen string) string {
	return util.GetPublicIP(pip, listen)

}
