package errorcode

type ErrorCode interface {
	Error() string
	Code() uint32
}

type ErrorWCode struct {
	err  string
	code uint32
}

func (ec *ErrorWCode) Error() string {
	return ec.err
}

func (ec *ErrorWCode) Code() uint32 {
	return ec.code
}

func New(errStr string, Code uint32) ErrorCode {
	ec := ErrorWCode{
		err:  errStr,
		code: Code,
	}

	return &ec
}

func NewByError(err error, Code uint32) ErrorCode {
	ec := ErrorWCode{
		err:  err.Error(),
		code: Code,
	}

	return &ec
}
