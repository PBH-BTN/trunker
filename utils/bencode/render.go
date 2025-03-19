package bencode

import (
	"io"

	"github.com/cloudwego/hertz/pkg/app/server/render"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cristalhq/bencode"
)

type BencodeRender struct {
	Data any
}

type FastBencode interface {
	Bencode(w io.Writer) error
}

func writeContentType(resp *protocol.Response, value string) {
	resp.Header.SetContentType(value)
}

var (
	_ render.Render = BencodeRender{}
)

func (r BencodeRender) Render(resp *protocol.Response) error {
	r.WriteContentType(resp)
	if t, ok := r.Data.(FastBencode); ok { // fast path
		return t.Bencode(resp.BodyWriter())
	}
	// fallback
	res, err := bencode.Marshal(r.Data)
	if err != nil {
		return err
	}
	resp.AppendBody(res)
	return nil
}

func (r BencodeRender) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, consts.MIMETextPlainUTF8)
}
