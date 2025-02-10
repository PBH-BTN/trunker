package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func LogSlowQuery(ctx context.Context, c *app.RequestContext) {
	now := time.Now()
	c.Next(ctx)
	cost := time.Since(now)
	if cost > time.Second*3 {
		hlog.CtxInfof(ctx, "slow query,req: %s ,resp:%s ,cost:%s", c.Request.RequestURI(), c.Response.Body(), cost.String())
	}
}
