package mux_local

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/DataDog/zstd"
	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/PBH-BTN/trunker/biz/services/peer/rpc"
	"github.com/PBH-BTN/trunker/kitex_gen/pbh/btn/trunker"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/cloudwego/frugal"
	"github.com/gofrs/flock"
)

func (m *MuxLocalManager) LoadFromPersist() {
	if !config.AppConfig.Tracker.Memory.EnablePersist {
		logger.Infof("persist not enabled, skip...")
		return
	}
	logger.Infof("start to load peers from persist")
	file, err := os.OpenFile(config.AppConfig.Tracker.Memory.PersistFile, os.O_RDONLY, 0644)
	if err != nil {
		logger.Errorf("open file error:%s", err.Error())
		return
	}
	reader := zstd.NewReader(bufio.NewReader(file))
	defer func() {
		_ = reader.Close()
		_ = file.Close()
	}()
	count := 0
	expired := 0
	now := time.Now()
	data := make([]byte, 0, 400)
	raminBuf := make([]byte, 0, 400)
	defer logger.Infof("load from persist done. %d peers loaded, %d peers expired", count, expired)
	var size uint32
	for {
		// Decode data length
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") { // end of file
				break
			}
			logger.Errorf("Failed to decode data length:%s", err.Error())
			return
		}
		data = data[:size]
		if readCount, err := reader.Read(data); err != nil {
			logger.Errorf("Failed to decode data length:%s", err.Error())
			return
		} else if uint32(readCount) != size {
			// read more
			remain := size - uint32(readCount)
			for remain > 0 {
				raminBuf = raminBuf[:remain]
				n, err := reader.Read(raminBuf)
				if err != nil {
					logger.Errorf("Failed to read data:%s", err.Error())
					return
				}
				remain -= uint32(n)
				data = append(data[0:readCount], raminBuf...)
			}
		}

		pbStruct := trunker.Store{}
		if _, err = frugal.DecodeObject(data[:size], &pbStruct); err != nil {
			logger.Errorf("Failed to decode peer:%s", err.Error())
			break
		}
		lastSeen := time.Unix(pbStruct.Peer.LastSeen, 0)
		if lastSeen.Add(time.Duration(config.AppConfig.Tracker.TTL) * time.Second).Before(now) {
			expired++
			continue
		}
		m.pickWorker(conv.UnsafeStringToBytes(pbStruct.InfoHash)).DirectStore(pbStruct.InfoHash, rpc.PeerIDLToCommon(pbStruct.Peer))
		count++
	}
}

func (m *MuxLocalManager) StoreToPersist() {
	if !config.AppConfig.Tracker.Memory.EnablePersist {
		logger.Infof("persist not enabled, skip...")
		return
	}
	tempFile := fmt.Sprintf("%s.%s", config.AppConfig.Tracker.Memory.PersistFile, rand.Text()[:8])
	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Errorf("open file error")
		return
	}
	writer := zstd.NewWriterLevel(bufio.NewWriter(file), 10)
	logger.Infof("start to store peers to persist")
	count := 0
	data := make([]byte, 800)
	for _, manager := range m.localList {
		manager.RangeMap(func(infoHash string, value *local.InfoHashRoot) bool {
			value.Range(func(key string, value *common.Peer) bool {
				obj := trunker.Store{
					InfoHash: infoHash,
					Peer:     rpc.PeerCommonToIDL(value),
				}
				n, err := frugal.EncodeObject(data, nil, obj)
				if err != nil {
					logger.Error("failed to marshal to pb:", err.Error())
					return true
				}
				// Encode data length
				if err := binary.Write(writer, binary.LittleEndian, uint32(n)); err != nil {
					logger.Error("Failed to encode data length:", err.Error())
					return true
				}
				if _, err := writer.Write(data[:n]); err != nil {
					logger.Error("Failed to write data:", err.Error())
					return false
				}
				count++
				return true
			})
			return true
		})
	}
	_ = writer.Flush()
	_ = writer.Close()
	err = file.Close()
	if err != nil {
		logger.Errorf("close file error:%s", err.Error())
		return
	}
	lock := flock.New(config.AppConfig.Tracker.Memory.PersistFile)
	if err := lock.Lock(); err != nil {
		logger.Errorf("failed to obtain write lock: %s", err.Error())
		_ = file.Close()
		return
	}
	defer func() {
		_ = lock.Unlock()
	}()
	err = os.Rename(tempFile, config.AppConfig.Tracker.Memory.PersistFile)
	if err != nil {
		logger.Errorf("failed to rename file from %s to %s: %s", tempFile, config.AppConfig.Tracker.Memory.PersistFile, err.Error())
		return
	}
	logger.Infof("store to persist done. %d peers stored", count)
}
