package interval

import (
	"time"

	"github.com/PBH-BTN/trunker/biz/config"
	"github.com/bytedance/gopkg/util/logger"
)

var (
	taskRunning = false
)

func StartIntervalTask() {
	go func() {
		logger.Infof("start interval task, interval %d seconds", config.AppConfig.Tracker.IntervalTask)
		for {
			<-time.After(time.Duration(config.AppConfig.Tracker.IntervalTask) * time.Second)
			doIntervalTask()
		}
	}()
}

var taskList = []func(){
	cleanInactivePeer,
	printStatics,
	saveDB,
}

func doIntervalTask() {
	if taskRunning {
		logger.Warn("task is running, skip")
		return
	}
	taskRunning = true
	for _, task := range taskList {
		task()
	}
	taskRunning = false
}
