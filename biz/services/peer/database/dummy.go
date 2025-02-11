package database

import "github.com/cloudwego/hertz/pkg/common/hlog"

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
