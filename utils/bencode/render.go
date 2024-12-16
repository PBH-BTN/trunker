package bencode

import (
	"github.com/cloudwego/hertz/pkg/app/server/render"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cristalhq/bencode"
)

type BencodeRender struct {
	Data any
}

const bencodeContentType = "text/plain; charset=utf-8"

func writeContentType(resp *protocol.Response, value string) {
	resp.Header.SetContentType(value)
}

var (
	_ render.Render = BencodeRender{}
)

func (r BencodeRender) Render(resp *protocol.Response) error {
	r.WriteContentType(resp)
	res, err := bencode.Marshal(r.Data)
	if err != nil {
		return err
	}
	resp.AppendBody(res)
	return nil
}

func (r BencodeRender) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, bencodeContentType)
}
