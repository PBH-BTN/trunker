package peer

import (
	"strconv"
	"time"

	"github.com/PBH-BTN/trunker/biz/model"
)

type Peer struct {
	model.Peer
	Uploaded   int       `json:"uploaded"`
	Downloaded int       `json:"downloaded"`
	LastSeen   time.Time `json:"lastSeen"`
}

func (p Peer) GetKey() string {
	return p.IP + strconv.FormatInt(int64(p.Port), 10)
}
