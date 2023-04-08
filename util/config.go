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

type AccessControl struct {
	AllowedUserID map[string]bool `yaml:"allowedUserID,omitempty"`
	AllowedChatID map[string]bool `yaml:"allowedChatID,omitempty"`
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

func ReadAccessList() AccessControl {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(exe)
	aclPath := filepath.Join(exPath, "acl.yml")

	aclFile, err := ioutil.ReadFile(aclPath)
	if err != nil {
		log.Error("fail to ready acl file %s, %v", aclPath, err)
		panic(err)
	}

	var acl AccessControl
	err = yaml.Unmarshal(aclFile, &acl)
	if err != nil {
		log.Error("fail to parse acl file %s, %v", aclPath, err)
		panic(err)
	}
	return acl
}
