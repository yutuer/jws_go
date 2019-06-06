package gm_command

//
// 暂时只能解析json
//

type Context struct {
	data string
}

func (c *Context) Clone() *Context {
	return &Context{}
}

func (c *Context) SetData(data string) {
	c.data = data
}

func (c *Context) GetData() string {
	return c.data
}

var context_p Context

func NewContext() *Context {
	return context_p.Clone()
}
