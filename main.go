/*
 * mastercoderk@gmail.com
 */

package main

import (
	"os"
	"path/filepath"

	"chloe/botservice"
	"chloe/util"

	log "github.com/jeanphorn/log4go"
)

func initLog() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(exe)
	logPath := filepath.Join(exPath, "log")
	// init log
	flw := log.NewFileLogWriter(filepath.Join(logPath, "chloe.log"), true, true)
	flw.SetFormat("[%D %T] [%L] (%S) %M")
	flw.SetRotate(true)
	flw.SetRotateDaily(true)
	flw.SetRotateSize(1024 * 1024 * 5)
	flw.SetRotateMaxBackup(5)
	log.AddFilter("DEFAULT", log.DEBUG, flw)
}

func main() {
	initLog()
	log.Info("openai bot Chloe Started.")

	config := util.ReadConfig()
	acl := util.ReadAccessList()
	service := botservice.NewTgBotService(config, acl)

	service.Run()
}
