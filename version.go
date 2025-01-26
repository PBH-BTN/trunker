package main

import (
	"log"
	"runtime"
)

var (
	Commit string
)

func init() {
	log.Printf("Runtime: %s\tCommit: %s \n", runtime.Version(), Commit)
}
