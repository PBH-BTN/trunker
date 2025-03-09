package main

import (
	"log"
	"runtime"
	"strconv"
	"time"
)

var (
	Commit         = "n/a"
	Version        = "dev"
	BuildTimestamp = ""
)

func init() {
	log.Printf("Version: %s\tCommit:%s\n", Version, Commit)
	if BuildTimestamp == "" {
		BuildTimestamp = "n/a"
	} else {
		ts, err := strconv.ParseInt(BuildTimestamp, 10, 64)
		if err == nil {
			t := time.Unix(ts, 0)
			BuildTimestamp = t.Format("2006-01-02 15:04:05 MST")
		}
	}
	log.Printf("Runtime: %s\tBuild Time: %s\n", runtime.Version(), BuildTimestamp)
}
