package pay

const (
	payTyp_IAP = iota
	payTyp_Count
)

type payImp interface {
	ProcessResData(string) error
}

var payImps [payTyp_Count]payImp

func init() {
	payImps[payTyp_IAP] = payImpIAP{}
}

func processPayData(typ int, data string) error {
	return nil
}
