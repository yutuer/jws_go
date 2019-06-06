package gin_core

import "html/template"

var html = template.Must(template.New("yyb_gift").Parse(`
{"ret":{{.ret}},"msg":"{{.message}}"}
`))
