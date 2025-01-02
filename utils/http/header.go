package http

import (
	"context"
	"net"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var ipHeader = []string{
	"X-Forwarded-For",
	"X-Real-IP",
}

func GetClientIP(ctx context.Context, c *app.RequestContext) net.IP {
	// There is a bug in c.ClientIP() while using Unix Domain Socket, unfortunately hertz don't want to fix this.
	// So we have to use this workaround.
	for _, header := range ipHeader {
		ip := c.Request.Header.Get(header)
		if ip != "" {
			ips := strings.Split(ip, ",")
			for _, ip := range ips {
				res := net.ParseIP(ip)
				if res != nil {
					return res
				}
			}
			hlog.CtxWarnf(ctx, "invalid ip from header %s:%s", header, ip)
		}
	}
	ip := c.ClientIP()
	if ip == "" {
		hlog.CtxWarnf(ctx, "failed to get client ip,header:%s", c.Request.Header.Header())
		return nil
	}
	return net.ParseIP(ip)
}
