package mux_local

import (
	"bufio"
	"encoding/binary"
	"os"
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/PBH-BTN/trunker/biz/services/peer/common"
	"github.com/PBH-BTN/trunker/biz/services/peer/local"
	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bytedance/gopkg/util/logger"
	"google.golang.org/protobuf/proto"
)

const PersistDataName = "persist.dat"

func (m *MuxLocalManager) LoadFromPersist() {
	if !config.AppConfig.Tracker.EnablePersist {
		logger.Info("persist not enabled, skip...")
		return
	}
	logger.Info("start to load peers from persist")
	file, err := os.OpenFile(PersistDataName, os.O_RDONLY, 0644)
	if err != nil {
		logger.Error("open file error:", err.Error())
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	count := 0
	expired := 0
	for {
		var size uint32
		// Decode data length
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			if err.Error() == "EOF" { // end of file
				break
			}
			logger.Error("Failed to decode data length:", err.Error())
			return
		}
		data := make([]byte, size)
		if readCount, err := reader.Read(data); err != nil {
			logger.Error("Failed to decode data length:", err.Error())
			return
		} else if uint32(readCount) != size {
			// read more
			remain := size - uint32(readCount)
			for remain > 0 {
				tmp := make([]byte, remain)
				n, err := reader.Read(tmp)
				if err != nil {
					logger.Error("Failed to decode data length:", err.Error())
					return
				}
				remain -= uint32(n)
				data = append(data[0:readCount], tmp...)
			}
		}

		// Unmarshal to protobuf SomeStruct
		pbStruct := &PeerInfo{}
		if err := proto.Unmarshal(data, pbStruct); err != nil {
			logger.Error("Failed to decode data length:", err.Error())
			break
		}
		lastSeen := time.Unix(pbStruct.LastSeen, 0)
		if lastSeen.Add(time.Duration(config.AppConfig.Tracker.TTL) * time.Second).Before(time.Now()) {
			// expired, skip
			expired++
			continue
		}
		m.pickWorker(pbStruct.InfoHash).DirectStore(string(pbStruct.InfoHash), &common.Peer{
			ID:         string(pbStruct.PeerId),
			IP:         pbStruct.Ip.ReportIp,
			IPv4:       pbStruct.Ip.ReportV4,
			IPv6:       pbStruct.Ip.ReportV6,
			ClientIP:   pbStruct.Ip.ClientIp,
			Port:       int(pbStruct.Port),
			Left:       pbStruct.Left,
			Uploaded:   pbStruct.Uploaded,
			Downloaded: pbStruct.Downloaded,
			LastSeen:   lastSeen,
			UserAgent:  pbStruct.UserAgent,
			Event:      common.PeerEvent(pbStruct.Event),
		})
		count++
	}
	logger.Infof("load from persist done. %d peers loaded,%d peers expired", count, expired)
}

func (m *MuxLocalManager) StoreToPersist() {
	if !config.AppConfig.Tracker.EnablePersist {
		logger.Info("persist not enabled, skip...")
		return
	}
	file, err := os.OpenFile(PersistDataName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("open file error")
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	logger.Info("start to store peers to persist")
	count := 0
	for _, manager := range m.localList {
		manager.RangeMap(func(infoHash string, value *local.InfoHashRoot) bool {
			value.Range(func(key string, value *common.Peer) bool {
				peerPB := &PeerInfo{
					PeerId:   conv.UnsafeStringToBytes(value.ID),
					InfoHash: conv.UnsafeStringToBytes(infoHash),
					Ip: &IPInfo{
						ClientIp: value.ClientIP,
						ReportIp: value.IP,
						ReportV4: value.IPv4,
						ReportV6: value.IPv6,
					},
					Port:       int32(value.Port),
					Left:       value.Left,
					Downloaded: value.Downloaded,
					Uploaded:   value.Uploaded,
					LastSeen:   value.LastSeen.Unix(),
					UserAgent:  value.UserAgent,
					Event:      PeerEvent(value.Event),
				}
				data, err := proto.Marshal(peerPB)
				if err != nil {
					logger.Error("failed to marshal to pb:", err.Error())
					return true
				}

				// Encode data length
				if err := binary.Write(writer, binary.LittleEndian, uint32(len(data))); err != nil {
					logger.Error("Failed to encode data length:", err.Error())
					return true
				}
				if _, err := writer.Write(data); err != nil {
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
	logger.Infof("store to persist done. %d peers stored", count)
}
