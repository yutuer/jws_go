package helper

import (
	"encoding/json"
)

//..
const (
	ProjectName = "jws"
	ServiceName = "crossservice"
)

//ServiceParam ..
type ServiceParam struct {
	GroupIDs []uint32 `json:"group_id,omitempty"`
	Gid      uint32   `json:"gid,omitempty"`

	IP   string `json:"ip,omitempty"`
	Port int    `json:"port,omitempty"`
}

//MarshalServiceParam ..
func MarshalServiceParam(p *ServiceParam) (string, error) {
	bs, err := json.Marshal(p)
	return string(bs), err
}

//UnmarshalServiceParam ..
func UnmarshalServiceParam(s string) (*ServiceParam, error) {
	p := &ServiceParam{}
	err := json.Unmarshal([]byte(s), p)
	return p, err
}
