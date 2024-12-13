package bencode

import (
	"github.com/PBH-BTN/trunker/biz/model"
	"github.com/cloudwego/hertz/pkg/app"
)

func ResponseOk(c *app.RequestContext, data any) {
	c.Render(200, BencodeRender{data})
}

func ResponseErr(c *app.RequestContext, err error) {
	c.Render(400, BencodeRender{model.ErrorResponse{FailureReason: err.Error(), Retry: "never"}})
}
