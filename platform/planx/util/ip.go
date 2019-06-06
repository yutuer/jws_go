package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"vcs.taiyouxi.net/platform/planx/util/logs"

	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cenk/backoff"
)

func GetPublicIP(pip, listen string) string {
	if _, p, err := net.SplitHostPort(listen); err != nil {
		panic(fmt.Sprintf("[GateServer] config listen has problem(aws). %s", err.Error()))
	} else {
		if _, _, err := net.SplitHostPort(pip); err == nil {
			return pip
		}
		switch pip {
		case "aws", "AWS":
			if awsip, e := AwsGetPublicIP(); e != nil {
				panic(fmt.Sprintf("[GateServer] config listen has problem(aws). %s", e.Error()))
			} else {
				return net.JoinHostPort(awsip, p)
			}
		case "qcloud", "QCLOUD":
			if awsip, e := QcloudGetPublicIP(); e != nil {
				panic(fmt.Sprintf("[GateServer] config listen has problem(aws). %s", e.Error()))
			} else {
				return net.JoinHostPort(awsip, p)
			}
		case "docker", "Docker", "DOCKER":
			if envListenIP := os.Getenv("HostPublicIP"); "" != envListenIP {
				if envPort8667 := os.Getenv("PORT_8667"); "" != envPort8667 {
					return net.JoinHostPort(envListenIP, envPort8667)
				}
				panic(fmt.Sprintf("[GateServer] config listen has problem(docker). no PORT_8667 environment"))
			}
			panic(fmt.Sprintf("[GateServer] config listen has problem(docker). no HostPublicIP environment"))
		default:
			host, err := ExternalIP()
			if err != nil {
				panic(fmt.Sprintf("[GateServer] config listen has problem(externalIP). %s", err.Error()))
			} else {
				return net.JoinHostPort(host, p)
			}
		}
	}

}

func QcloudGetPublicIP() (string, error) {
	var ret string
	operation := func() error {
		response, err := http.Get("http://metadata.tencentyun.com/meta-data/public-ipv4")
		if err != nil {
			logs.Error("QcloudGetPublicIP get ipv4 failed. %s", err.Error())
			return err
		} else {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			ret = string(body)
			if err != nil {
				logs.Error("QcloudGetPublicIP get ipv4 failed. %s", err.Error())
				return err
			}
			return nil
		}
		return fmt.Errorf("QcloudGetPublicIP should not happen") // or an error
	}
	ebo := backoff.NewExponentialBackOff()
	ebo.MaxElapsedTime = 10 * time.Second
	err := backoff.Retry(operation, ebo)
	if err != nil {
		// Handle error.
		logs.Error("QcloudGetPublicIP get ipv4 finally failed. %s", err.Error())
		return "", err
	}
	return ret, nil
}

func AwsGetPublicIP() (string, error) {
	sessDefaults := session.New()
	meta := ec2metadata.New(sessDefaults)
	ip, err := meta.GetMetadata("public-ipv4")
	if err != nil {
		return "", err
	}
	return ip, nil
}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
