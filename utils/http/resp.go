package http

import "github.com/cloudwego/hertz/pkg/app"

type commonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func ResponseBadRequest(c *app.RequestContext) {
	c.JSON(400, commonResponse{
		Code:    400,
		Message: "Bad Request",
	})
}

func ResponseUnauthorized(c *app.RequestContext) {
	c.JSON(401, commonResponse{
		Code:    401,
		Message: "Unauthorized",
	})
}

func ResponseOK(c *app.RequestContext, data any) {
	c.JSON(200, commonResponse{
		Code:    200,
		Message: "ok",
		Data:    data,
	})
}
