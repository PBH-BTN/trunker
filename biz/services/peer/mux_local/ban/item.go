package ban

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/PBH-BTN/trunker/utils/conv"
	"github.com/bits-and-blooms/bloom/v3"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type banItem struct {
	m      sync.RWMutex
	filter *bloom.BloomFilter
	cap    uint
	count  uint
	store  *os.File
}

const defaultFilterCap = 32

func newBanItem(storeName string) (*banItem, error) {
	store, err := os.CreateTemp("", storeName)
	if err != nil {
		return nil, err
	}
	return &banItem{
		m:      sync.RWMutex{},
		filter: bloom.NewWithEstimates(defaultFilterCap, 0.01),
		cap:    defaultFilterCap,
		count:  0,
		store:  store,
	}, nil
}

func (i *banItem) test(target []byte) bool {
	i.m.RLock()
	defer i.m.RUnlock()
	return i.filter.Test(target)
}

func (i *banItem) clear() {
	i.m.Lock()
	defer i.m.Unlock()
	i.filter = bloom.NewWithEstimates(defaultFilterCap, 0.01)
	i.cap = defaultFilterCap
	i.count = 0
	// clear file
	_ = i.store.Truncate(0)
	_, _ = i.store.Seek(0, io.SeekStart)
}

func (i *banItem) add(target string) error {
	i.m.Lock()
	defer i.m.Unlock()
	if i.count+1 >= i.cap {
		// add with grow
		i.cap *= 2                         // double the cap
		currentBanlist, err := i.readBan() // read current ban list
		if err != nil {
			hlog.Errorf("read ban list error: %s", err.Error())
			return err
		}
		newFilter := bloom.NewWithEstimates(i.cap, 0.01)
		for _, v := range currentBanlist {
			newFilter.Add(conv.UnsafeStringToBytes(v)) // migrate old ban list
		}
		newFilter.Add(conv.UnsafeStringToBytes(target)) // add current
		i.filter = newFilter
	} else {
		i.filter.Add(conv.UnsafeStringToBytes(target))
	}
	i.count++
	_, err := i.store.WriteString(target + "\n")
	if err != nil {
		hlog.Errorf("write ban list error: %s", err.Error())
		return err
	}
	return i.store.Sync()
}

func (i *banItem) readBan() ([]string, error) {
	var keys []string
	_, err := i.store.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(i.store)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}
