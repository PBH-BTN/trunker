package interval

import "github.com/PBH-BTN/trunker/biz/services/peer"

func cleanInactivePeer() {
	peer.GetAllMap().Range(func(key string, value *peer.InfoHashRoot) bool {
		peer.CleanUp(value)
		return true
	})
}

func saveDB() {
	peer.SavePeer()
}
