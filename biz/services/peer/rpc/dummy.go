package rpc

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func (m Manager) Clean() int64 {
	hlog.Info("no need for rpc mode, skip")
	return 0
}

func (m Manager) LoadFromPersist() {
	hlog.Info("no need for rpc mode, skip")
}

func (m Manager) StoreToPersist() {
	hlog.Info("no need for rpc mode, skip")
}

func (m Manager) AnswerToPeer(ctx context.Context, infoHash string, peerID string, answerBody []byte) error {
	return errors.New("not supported")
}
