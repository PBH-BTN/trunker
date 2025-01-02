package http

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func TestGetClientIP(t *testing.T) {
	type args struct {
		c *app.RequestContext
	}
	header1 := protocol.RequestHeader{}
	header1.Set("X-Forwarded-For", "192.168.1.1")
	header2 := protocol.RequestHeader{}
	header2.Set("X-Real-IP", "2a09:bac5:2431:e6::17:358")
	header3 := protocol.RequestHeader{}
	header3.Set("X-Forwarded-For", "2a09:bac5:2431:e6::17:358,2a09:bac5:2431:e6::17:358")
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{
			"ipv4",
			args{
				&app.RequestContext{
					Request: protocol.Request{
						Header: header1,
					},
				},
			},
			net.IPv4(192, 168, 1, 1),
		},
		{
			"ipv6",
			args{
				&app.RequestContext{
					Request: protocol.Request{
						Header: header2,
					},
				},
			},
			net.ParseIP("2a09:bac5:2431:e6::17:358"),
		},
		{
			"multi ipv6",
			args{
				&app.RequestContext{
					Request: protocol.Request{
						Header: header3,
					},
				},
			},
			net.ParseIP("2a09:bac5:2431:e6::17:358"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetClientIP(context.Background(), tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
