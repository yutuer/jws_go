package exclusion

import (
	"log"
	"testing"
)

func TestExclusion(t *testing.T) {
	handle := NewHandle("TestExclusion")

	handle.AddNode("aaad", "aaad")
	handle.AddNode("aaad2", "aaad2")
	handle.AddNode("aaad", "aaad")

	log.Printf("Nodes: %+v", handle.Nodes)

	handle.UnPublish()
}
