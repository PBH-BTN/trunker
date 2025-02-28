package database

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func (m *DBManager) Clean() int64 {
	hlog.Info("no need for database mode, skip")
	return 0
}

func (m *DBManager) LoadFromPersist() {
	hlog.Info("no need for database mode, skip")
}

func (m *DBManager) StoreToPersist() {
	hlog.Info("no need for database mode, skip")
}

func (m *DBManager) AnswerToPeer(_ context.Context, _ string, _ string, _ []byte) error {
	hlog.Warn("AnswerToPeer: peer communication not supported in database mode")
	return nil
}
