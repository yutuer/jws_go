package discover

import (
	"encoding/json"
	"strconv"
	"strings"
)

//Service ..
type Service struct {
	Project string
	Version string
	Build   string
	Service string
	Index   string
	IP      string
	Port    int
	Extra   string
}

//ParseVersion major.minor.fix.description
func (s Service) ParseVersion() (int, int, int, int) {
	arr := strings.Split(s.Version, ".")
	return atoiSafe(arr, 0), atoiSafe(arr, 1), atoiSafe(arr, 2), atoiSafe(arr, 3)
}

func atoiSafe(s []string, i int) int {
	if i >= len(s) {
		return 0
	}
	ri, err := strconv.Atoi(s[i])
	if nil != err {
		return 0
	}
	return ri
}
func encodeValue(s *Service) (string, error) {
	bs, err := json.Marshal(s)
	if nil != err {
		return "", err
	}
	return string(bs), nil
}

func decodeValue(str string) (*Service, error) {
	s := &Service{}
	err := json.Unmarshal([]byte(str), s)
	if nil != err {
		return nil, err
	}
	return s, nil
}
