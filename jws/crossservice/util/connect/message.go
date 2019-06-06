package connect

//Message ..
type Message struct {
	Payload []byte
	Length  int
}

//NewMessage ..
func NewMessage(payload []byte, length int) *Message {
	msg := &Message{
		Payload: payload[:length],
		Length:  length,
	}

	return msg
}

type packet struct {
	data   []byte
	length int
}
