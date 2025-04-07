package middleware

import (
	"context"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func LogSlowQuery() app.HandlerFunc {
	slowTime := utils.If(config.AppConfig.Tracker.Mode == config.RunningModeMemory, time.Millisecond*100, time.Second*3)
	return func(ctx context.Context, c *app.RequestContext) {
		now := time.Now()
		c.Next(ctx)
		cost := time.Since(now)
		if cost > slowTime {
			hlog.CtxInfof(ctx, "slow query,req: %s ,resp: %s ,cost: %s", c.Request.RequestURI(), c.Response.Body(), cost.String())
		}
	}
}
