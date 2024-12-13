package peer

import (
	"io"
	"os"

	"github.com/PBH-BTN/trunker/utils"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/shamaton/msgpack/v2"
	"github.com/zhangyunhao116/skipmap"
)

type dbFormat struct {
	InfoHash map[string]*infoHash
}
type infoHash struct {
	Peer []*Peer
}

func SavePeer() {
	f, err := os.OpenFile("peer.db", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("failed to save db:%s", err.Error())
		return
	}
	defer f.Close()
	raw, err := msgpack.Marshal(buildDBHash())
	if err != nil {
		logger.Errorf("failed to save db:%s", err.Error())
		return
	}
	_, err = f.Write(raw)
	if err != nil {
		logger.Errorf("failed to save db:%s", err.Error())
		return
	}
}

func LoadPeer() {
	f, err := os.OpenFile("peer.db", os.O_RDONLY, 0644)
	if err != nil {
		logger.Errorf("failed to read from db:%s", err.Error())
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		logger.Errorf("failed to read from db:%s", err.Error())
		return
	}
	var db *dbFormat
	err = msgpack.Unmarshal(raw, db)
	if err != nil {
		logger.Errorf("failed to read from db:%s", err.Error())
		return
	}
	rebuildFromDB(db)
}

func buildDBHash() *dbFormat {
	ret := &dbFormat{InfoHash: make(map[string]*infoHash)}
	infoHashMap.Range(func(key string, value *InfoHashRoot) bool {
		ret.InfoHash[key] = &infoHash{Peer: utils.SkipMapToSlice(value.peerMap)}
		return true
	})
	return ret
}

func rebuildFromDB(db *dbFormat) {
	for k, v := range db.InfoHash {
		root := &InfoHashRoot{peerMap: skipmap.New[string, *Peer]()}
		for _, peer := range v.Peer {
			root.peerMap.Store(peer.GetKey(), peer)
		}
		infoHashMap.Store(k, root)
	}
	return
}
