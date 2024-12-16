package middleware

import (
	"context"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

func LogSlowQuery(ctx context.Context, c *app.RequestContext) {
	now := time.Now()
	c.Next(ctx)
	cost := time.Since(now)
	if cost > time.Second*5 {
		logger.CtxInfof(ctx, "slow query,req: %s ,resp:%s ,cost:%s", c.Request.RequestURI(), c.Response.Body(), cost.String())
	}
}
