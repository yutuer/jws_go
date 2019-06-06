package connect

import (
	"log"
	"testing"
)

func TestFilter(t *testing.T) {
	filter := NewIPFilter()
	filter.Add("10.222.0.0/24")

	log.Printf("-----%v", filter.check("10.222.0.105:2222"))
}
