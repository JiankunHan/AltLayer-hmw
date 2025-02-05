package utils

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// read from config.yml
type Config struct {
	Queue struct {
		BufferSize int `yaml:"buffer_size"`
	} `yaml:"queue"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	ReqHandler struct {
		NumReqHandlers int `yaml:"num_req_handlers"`
	} `yaml:"req_handlers"`
	Transaction struct {
		MaxRetryTimes int `yaml:"max_retries"`
	} `yaml:"transaction"`
}

func LoadConfig(filename string) (*Config, error) {
	// 读取配置文件内容
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// 解析 YAML 配置
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
