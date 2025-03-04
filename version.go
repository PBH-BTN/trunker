package main

import (
	"log"
	"runtime"
)

var (
	Commit         = "n/a"
	Version        = "dev"
	BuildTimestamp = "n/a"
)

func init() {
	log.Printf("Version: %s\tCommit:%s\n", Version, Commit)
	log.Printf("Runtime: %s\tBuild Time: %s\n", runtime.Version(), BuildTimestamp)
}
