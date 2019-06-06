package distinct

import (
	"testing"
	"fmt"
)
type trying struct {
	a int
	b int
}
func TestDisinct(t *testing.T){
	info := []uint32{5,6,55,5,7,5,8,9,9,9}


	result,err := ValuesAndDisinct(info)
	fmt.Printf("%v  err:%v",result,err)
}
