package middleware

import (
	"context"
	"os"

	"github.com/PBH-BTN/trunker/utils/http"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

func GetAdminAuthMiddleware() []app.HandlerFunc {
	key := os.Getenv("ADMIN_KEY")
	if key != "" {
		expectedHeader := "Bearer " + key
		return []app.HandlerFunc{func(c context.Context, ctx *app.RequestContext) {
			if ctx.Request.Header.Get("Authorization") != expectedHeader {
				http.ResponseUnauthorized(ctx)
				ctx.Abort()
				return
			}
			ctx.Next(c)
		}}
	} else {
		logger.Warn("admin key is not set, please set ADMIN_KEY environment variable!")
		return nil
	}
}
