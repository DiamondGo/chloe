/*
 * mastercoderk@gmail.com
 */

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/jeanphorn/log4go"
	"gopkg.in/yaml.v3"
)

type Config struct {
	BotName string `yaml:"botName"`
	OpenAI  struct {
		APIKey         string `yaml:"apiKey"`
		Model          string `yaml:"model"`
		ContextTimeout int    `yaml:"contextTimeout"`
	} `yaml:"openAI"`
	Telegram struct {
		BotToken string `yaml:"botToken"`
	} `yaml:"telegram"`
	System struct {
		WhitelistEnabled bool `yaml:"whitelistEnabled"`
	} `yaml:"system"`
}

func ReadConfig() Config {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(exe)
	configPath := filepath.Join(exPath, "config.yml")

	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Error("fail to ready config file %s, %v", configPath, err)
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Error("fail to parse config file %s, %v", configPath, err)
		panic(err)
	}

	return config
}
