package local

import (
	"io"
	"os"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/utils"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/shamaton/msgpack/v2"
	"github.com/zhangyunhao116/skipmap"
)

type dbFormat struct {
	InfoHash map[string]*infoHash
}
type infoHash struct {
	Peer []*common.Peer
}

func (m *Manager) StoreToPersist() {
	if !config.AppConfig.Tracker.EnablePersist {
		logger.Info("persist not enabled, skip...")
		return
	}
	f, err := os.OpenFile("peer.db", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("failed to save db:%s", err.Error())
		return
	}
	defer f.Close()
	ret := &dbFormat{InfoHash: make(map[string]*infoHash)}
	m.infoHashMap.Range(func(key string, value *infoHashRoot) bool {
		ret.InfoHash[key] = &infoHash{Peer: utils.SkipMapToSlice(value.peerMap)}
		return true
	})
	raw, err := msgpack.Marshal(ret)
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

func (m *Manager) LoadFromPersist() {
	if !config.AppConfig.Tracker.EnablePersist {
		logger.Info("persist not enabled, skip...")
	}
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
	var db dbFormat
	err = msgpack.Unmarshal(raw, &db)
	if err != nil {
		logger.Errorf("failed to read from db:%s", err.Error())
		return
	}
	for k, v := range db.InfoHash {
		root := &infoHashRoot{peerMap: skipmap.New[string, *common.Peer]()}
		for _, peer := range v.Peer {
			root.peerMap.Store(peer.GetKey(), peer)
		}
		m.peerCount.Add(int64(len(v.Peer)))
		m.infoHashMap.Store(k, root)
	}
}
