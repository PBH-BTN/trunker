package config

import (
	"io"
	"log"
	"os"

	json "github.com/bytedance/sonic"

	"gopkg.in/yaml.v3"
)

var AppConfig *Config

type databaseConfig struct {
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
	Enable  bool   `yaml:"enable" json:"enable"`
	Network string `yaml:"network" json:"network"`
	Addr    string `yaml:"addr" json:"addr"`
}

type runningMode string

const (
	RunningModeMemory runningMode = "memory"
	RunningModeDB     runningMode = "db"
	RunningModeRPC    runningMode = "rpc"
)

type TrackerConfig struct {
	Database            databaseConfig `yaml:"database" json:"database"`
	Memory              memoryConfig   `yaml:"memory" json:"memory"`
	UDPServer           udpConfig      `yaml:"udpServer" json:"udpServer"`
	WSServer            wsConfig       `yaml:"wsServer" json:"wsServer"`
	Mode                runningMode    `yaml:"mode" json:"mode"`
	RPC                 rpcConfig      `yaml:"rpc" json:"rpc"`
	XDP                 xdpConfig      `yaml:"xdp" json:"xdp"`
	HostPorts           string         `yaml:"hostPorts" json:"hostPorts"`
	MetricsHostPorts    string         `yaml:"metricsHostPorts" json:"metricsHostPorts"`
	TTL                 int64          `yaml:"ttl" json:"ttl"`
	IntervalTask        int64          `yaml:"intervalTask" json:"intervalTask"`
	UseUnixSocket       bool           `yaml:"useUnixSocket" json:"useUnixSocket"`
	UseAnnounceIP       bool           `yaml:"useAnnounceIP" json:"useAnnounceIP"` // allow peer to announce it external ip
	EnableEventProducer bool           `yaml:"enableEventProducer" json:"enableEventProducer"`
	EnableMetrics       bool           `yaml:"enableMetrics" json:"enableMetrics"`
	TrackerId           string         `yaml:"trackerId" json:"trackerId"`
	DebugPort           int64          `yaml:"debugPort" json:"debugPort"`
	RecordUserAgent     bool           `yaml:"recordUserAgent" json:"recordUserAgent"`
}

type wsConfig struct {
	Enable bool `yaml:"enable" json:"enable"`
	Shard  int  `yaml:"shard" json:"shard"`
}

type rpcConfig struct {
	EnableServer bool   `yaml:"enableServer" json:"enableServer"`
	HostPorts    string `yaml:"hostPorts" json:"hostPorts"`
	RemoteServer string `yaml:"remoteServer" json:"remoteServer"`
}

type udpConfig struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	HostPorts string `yaml:"hostPorts" json:"hostPorts"`
}

type memoryConfig struct {
	PersistFile        string `yaml:"persistFile" json:"persistFile"`
	MaxPeersPerTorrent int    `yaml:"maxPeersPerTorrent" json:"maxPeersPerTorrent"`
	Shard              int    `yaml:"shard" json:"shard"`
	EnablePersist      bool   `yaml:"enablePersist" json:"enablePersist"`
}

type xdpConfig struct {
	// Enable XDP-based IP filtering (Linux only)
	Enable bool `yaml:"enable" json:"enable"`
	// Interface is the network interface name to attach XDP program (e.g., "eth0", "ens33")
	// If empty, XDP will be disabled even if Enable is true
	Interface string `yaml:"interface" json:"interface"`
}

type Config struct {
	Cache    RedisConfig    `yaml:"cache" json:"cache"`
	Tracker  TrackerConfig  `yaml:"tracker" json:"tracker"`
	RocketMq RocketMqConfig `yaml:"rocketmq" json:"rocketmq"`
}

func injectDefaultValue(conf *Config) {
	if conf.Tracker.HostPorts == "" {
		conf.Tracker.HostPorts = "0.0.0.0:8888"
	}
	if conf.Tracker.Mode == RunningModeMemory {
		if conf.Tracker.Memory.PersistFile == "" {
			conf.Tracker.Memory.PersistFile = "persist.dat"
		}
	}
	if conf.Tracker.Mode == RunningModeRPC {
		if conf.Tracker.RPC.RemoteServer == "" {
			panic("rpc remote server is required")
		}
	}
	if conf.Tracker.MetricsHostPorts == "" {
		conf.Tracker.MetricsHostPorts = "127.0.0.1:9091"
	}
	if conf.Tracker.TrackerId == "" {
		conf.Tracker.TrackerId = "default"
	}
}
func Init() {
	config := &Config{}
	if jsonConfig := os.Getenv("TRUNKER_CONFIG"); jsonConfig != "" {
		log.Print("using json config from env")
		err := json.UnmarshalString(jsonConfig, config)
		if err != nil {
			panic("invalid json config:" + err.Error())
		}
		injectDefaultValue(config)
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
	injectDefaultValue(config)
	AppConfig = config
}
