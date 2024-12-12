package bencode

import "github.com/cloudwego/hertz/pkg/app"

func ResponseOk(c *app.RequestContext, data any) {
	c.Render(200, BencodeRender{data})
}
