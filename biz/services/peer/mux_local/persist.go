package mux_local

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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
	"github.com/xxjwxc/gowp/workpool"
	"go.uber.org/zap/buffer"
)

const writeFileThread = 10

func (m *MuxLocalManager) LoadFromPersist() {
	if !config.AppConfig.Tracker.Memory.EnablePersist {
		logger.Infof("persist not enabled, skip...")
		return
	}
	bufferPool := buffer.NewPool()
	files, _ := filepath.Glob(config.AppConfig.Tracker.Memory.PersistFile + ".*")
	count := atomic.Int64{}
	expired := atomic.Int64{}
	logger.Infof("start to load peers from persist, shard %d", len(files))
	for _, fileName := range files {
		go func() {
			lock := flock.New(fileName)
			ok, err := lock.TryRLock()
			if err != nil {
				logger.Errorf("failed to obtain read lock: %s", err.Error())
				return
			}
			if !ok {
				logger.Warnf("file %s is writing, skip", fileName)
			}
			defer func() {
				_ = lock.Unlock()
			}()
			file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
			if err != nil {
				logger.Errorf("open file error:%s", err.Error())
				return
			}
			reader := zstd.NewReader(bufio.NewReader(file))
			defer func() {
				_ = reader.Close()
				_ = file.Close()
			}()

			now := time.Now()
			dataBuf := bufferPool.Get()
			defer dataBuf.Free()
			data := dataBuf.Bytes()
			raminBuffer := bufferPool.Get()
			defer raminBuffer.Free()
			raminBuf := raminBuffer.Bytes()
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
					expired.Add(1)
					continue
				}
				m.pickWorker(conv.UnsafeStringToBytes(pbStruct.InfoHash)).DirectStore(pbStruct.InfoHash, rpc.PeerIDLToCommon(pbStruct.Peer))
				count.Add(1)
			}
		}()

	}
	logger.Infof("load from persist done. %d peers loaded, %d peers expired", count.Load(), expired.Load())

}

func (m *MuxLocalManager) StoreToPersist() {
	if !config.AppConfig.Tracker.Memory.EnablePersist {
		logger.Infof("persist not enabled, skip...")
		return
	}
	logger.Infof("start to store peers to persist")
	count := atomic.Int64{}
	bufferPool := buffer.NewPool()
	wp := workpool.New(writeFileThread)
	for i, manager := range m.localList {
		fileName := fmt.Sprintf("%s.%d", config.AppConfig.Tracker.Memory.PersistFile, i)
		wp.Do(func() error {
			lock := flock.New(fileName)
			if err := lock.Lock(); err != nil {
				logger.Errorf("failed to obtain write lock: %s", err.Error())
				return err
			}
			defer func() {
				_ = lock.Unlock()
			}()
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logger.Error("open file error:", err.Error())
				return err
			}
			writer := zstd.NewWriterLevel(bufio.NewWriter(file), 10)
			manager.RangeMap(func(infoHash string, value *local.InfoHashRoot) bool {
				value.Range(func(key string, value *common.Peer) bool {
					obj := trunker.Store{
						InfoHash: infoHash,
						Peer:     rpc.PeerCommonToIDL(value),
					}
					var n int
					buf := bufferPool.Get()
					data := buf.Bytes()
					defer buf.Free()
					n, err = frugal.EncodeObject(data, nil, obj)
					if err != nil {
						logger.Error("failed to marshal to pb:", err.Error())
						return true
					}
					// Encode data length
					if err = binary.Write(writer, binary.LittleEndian, uint32(n)); err != nil {
						logger.Error("Failed to encode data length:", err.Error())
						return true
					}
					if _, err = writer.Write(data[:n]); err != nil {
						logger.Error("Failed to write data:", err.Error())
						return false
					}
					count.Add(1)
					return true
				})
				return true
			})
			_ = writer.Flush()
			_ = writer.Close()
			err = file.Close()
			if err != nil {
				logger.Errorf("close file error:%s", err.Error())
			}
			return err
		})
	}
	if err := wp.Wait(); err != nil {
		logger.Error("store to persist error:", err.Error())
		return
	}

	logger.Infof("store to persist done. %d peers stored", count.Load())
}
