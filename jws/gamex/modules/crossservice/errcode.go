package crossservice

//ErrCode ..
const (
	ErrOK = int(iota) + 200
	ErrInvalid
	ErrNotAlive
	ErrRemote
)
