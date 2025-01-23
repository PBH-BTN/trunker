package config

import (
	"io"
	"log"
	"os"

	json "github.com/bytedance/sonic"

	"gopkg.in/yaml.v3"
)

var AppConfig *Config

type MysqlConfig struct {
	Database string `yaml:"database" json:"database"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Pass     string `yaml:"pass" json:"pass"`
	User     string `yaml:"user" json:"user"`
}

type RocketMqConfig struct {
	Topic    string `yaml:"topic" json:"topic"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

type RedisConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

type TrackerConfig struct {
	TTL                 int64  `yaml:"ttl" json:"ttl"`
	IntervalTask        int64  `yaml:"intervalTask" json:"intervalTask"`
	UseDB               bool   `yaml:"useDB" json:"useDB"`
	EnablePersist       bool   `yaml:"enablePersist" json:"enablePersist"`
	PersistFile         string `yaml:"persistFile" json:"persistFile"`
	MaxPeersPerTorrent  int    `yaml:"maxPeersPerTorrent" json:"maxPeersPerTorrent"`
	Shard               int    `yaml:"shard" json:"shard"`
	UseUnixSocket       bool   `yaml:"useUnixSocket" json:"useUnixSocket"`
	HostPorts           string `yaml:"hostPorts" json:"hostPorts"`
	UseAnnounceIP       bool   `yaml:"useAnnounceIP" json:"useAnnounceIP"` // allow peer to announce it external ip
	EnableEventProducer bool   `yaml:"enableEventProducer" json:"enableEventProducer"`
	EnableMetrics       bool   `yaml:"enableMetrics" json:"enableMetrics"`
}

type Config struct {
	PersistDatabase MysqlConfig    `yaml:"database" json:"database"`
	Cache           RedisConfig    `yaml:"cache" json:"cache"`
	Tracker         TrackerConfig  `yaml:"tracker" json:"tracker"`
	RocketMq        RocketMqConfig `yaml:"rocketmq" json:"rocketmq"`
}

func Init() {
	config := &Config{}
	if jsonConfig := os.Getenv("TRUNKER_CONFIG"); jsonConfig != "" {
		log.Print("using json config from env")
		err := json.UnmarshalString(jsonConfig, config)
		if err != nil {
			panic("invalid json config:" + err.Error())
		}
		AppConfig = config
		return
	}
	confFile := "conf/local.yaml"
	if os.Getenv("RUN_ENV") == "prod" {
		confFile = "conf/prod.yaml"
	}
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
