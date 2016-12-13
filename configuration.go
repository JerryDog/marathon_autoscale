package main

import (
	"encoding/json"
	"io/ioutil"
)

type MarathonCfg struct {
	// 序列化配置文件
	MarathonUrl string
	DBName      string
	DBUser      string
	DBPass      string
	DBHost      string
	DBPort      string
	LogPath     string
}

type Configuration struct {
	// AutoScale configuration
	Marathon MarathonCfg
}

func (config *Configuration) FromFile(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(content, &config)
}

func FromFile(filePath string) (Configuration, error) {
	conf := &Configuration{}
	err := conf.FromFile(filePath)
	return *conf, err
}
