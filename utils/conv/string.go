package conv

import (
	"encoding/hex"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/text/encoding/charmap"
)

func TransUTF8To8859_1(raw []byte) []byte {
	encoded, err := charmap.ISO8859_1.NewEncoder().Bytes(raw)
	if err != nil {
		hlog.Warnf("failed to convert utf8 to 8859-1:%s,raw:%s", err.Error(), hex.EncodeToString(raw))
	}
	return encoded
}

func Trans8859_1ToUTF8(raw []byte) []byte {
	encoded, err := charmap.ISO8859_1.NewDecoder().Bytes(raw)
	if err != nil {
		hlog.Warnf("failed to convert 8859-1 to utf8:%s,raw:%s", err.Error(), hex.EncodeToString(raw))
	}
	return encoded
}
