package safetable

import "testing"
import "fmt"

type playerCache struct {
	Acid      string
	Name      string
	GuildName string
	GuildPos  int
}

func TestSafeTable(t *testing.T) {
	st := NewSafeTable(128)

	acid := "0:11:71f9f21a-c0c9-4e7b-aa33-72463320670f"

	pc := playerCache{Acid: acid, Name: "name"}
	st.Set(acid, pc)

	rpc := st.Get(acid).(playerCache)

	fmt.Println(rpc)
}
