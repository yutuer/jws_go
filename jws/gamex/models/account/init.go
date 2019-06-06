package account

import (
	"vcs.taiyouxi.net/jws/gamex/models/codec"
)

func decode(raw []byte, out interface{}) {
	codec.Decode(raw, out)
}

func encode(value interface{}) []byte {
	return codec.Encode(value)
}
