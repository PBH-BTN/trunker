package middleware

import (
	"context"
	"errors"

	"github.com/PBH-BTN/trunker/utils/bencode"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/limiter"
)

// AdaptiveLimit CPU sampling algorithm using BBR
func AdaptiveLimit(opts ...limiter.Option) app.HandlerFunc {
	l := limiter.NewLimiter(opts...)
	return func(c context.Context, ctx *app.RequestContext) {
		done, err := l.Allow()
		if err != nil {
			ctx.AbortWithError(consts.StatusTooManyRequests, err)
			bencode.ResponseErrWithRetry(ctx, errors.New("too many request"))
		} else {
			ctx.Next(c)
			done()
		}
	}
}
