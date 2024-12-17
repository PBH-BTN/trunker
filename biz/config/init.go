package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var AppConfig *Config

type MysqlConfig struct {
	Database string `yaml:"database"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Pass     string `yaml:"pass"`
	User     string `yaml:"user"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type TrackerConfig struct {
	TTL                int64  `yaml:"ttl"`
	IntervalTask       int64  `yaml:"intervalTask"`
	UseDB              bool   `yaml:"useDB"`
	EnablePersist      bool   `yaml:"enablePersist"`
	MaxPeersPerTorrent int    `yaml:"maxPeersPerTorrent"`
	Shard              int    `yaml:"shard"`
	UseUnixSocket      bool   `yaml:"useUnixSocket"`
	HostPorts          string `yaml:"hostPorts"`
	UseAnnounceIP      bool   `yaml:"useAnnounceIP"` // allow peer to announce it external ip
}

type Config struct {
	PersistDatabase MysqlConfig   `yaml:"database"`
	Cache           RedisConfig   `yaml:"cache"`
	Tracker         TrackerConfig `yaml:"tracker"`
}

func Init() {
	confFile := "conf/local.yaml"
	if os.Getenv("RUN_ENV") == "prod" {
		confFile = "conf/prod.yaml"
	}
	config := &Config{}
	fp, err := os.Open(confFile)
	if err != nil {
		log.Fatal(err)
	}
	content, err := io.ReadAll(fp)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(content, config); err != nil {
		log.Fatalf("parse local config failed: %v", err)
	}
	AppConfig = config
}
