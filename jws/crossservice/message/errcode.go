package message

//error code
const (
	ErrCodeOK = iota
	ErrCodeInvalid
	ErrCodeInner
	ErrCodeTimeout
	ErrCodeUnknownMethod
	ErrCodeEncode
	ErrCodeDecode
	ErrCodeClosed
)
