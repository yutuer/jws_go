package proto

import "github.com/google/flatbuffers/go"

func GetNewBuilder() *flatbuffers.Builder {
	return flatbuffers.NewBuilder(20480)
}
