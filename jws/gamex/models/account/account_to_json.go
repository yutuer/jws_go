package account

import "encoding/json"

func (p *Account) ToJSON() ([]byte, error) {
	return json.Marshal(*p)
}
