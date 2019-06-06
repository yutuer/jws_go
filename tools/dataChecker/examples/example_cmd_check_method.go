package examples

import (
	"flag"
)

/*
	CMD sample:
	go run main.go -mode=clonly
	go run main.go -mode=clonly -cl="Test.xlsx"
*/

func main() {
	modeFlag := flag.String("mode", "clonly",
		`The running mode of the tool: clonly
		clonly: Only run checklist`,
	)
	// 详见 https://golang.org/pkg/flag/
	checklistFlag := flag.String("cl", "Checklist.xlsx", "The filename of the checklist. e.g. Checklist.xlsx ")
	flag.Parse()

	switch *modeFlag {
	case "clonly":
		RunSomething(*checklistFlag)
	default:
		panic("Run mode error!!!")
	}
}

func RunSomething(flag string) {

}
