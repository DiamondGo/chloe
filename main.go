/*
 * mastercoderk@gmail.com
 */

package main

import (
	"flag"
	"os"
	"path/filepath"

	"chloe/ai"
	"chloe/botservice"

	log "github.com/jeanphorn/log4go"
	"github.com/sashabaranov/go-openai"
)

func initLog() {
	// init app
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

	tgBotToken := flag.String("tgtoken", "", "telegram bot token")
	botName := flag.String("name", "Chloe", "ai bot name")
	model := flag.String("model", openai.GPT3Dot5Turbo, "openai model")
	apiKey := flag.String("aikey", "", "openai api key")
	contextTimeout := flag.Int("contextTimeout", 60, "context awareness timeout in seconds")

	flag.Parse()
	if *tgBotToken == "" || *botName == "" || *model == "" || *apiKey == "" {
		panic("")
	}

	openaiConfig := ai.AIConfig{
		BotName:        *botName,
		Model:          *model,
		ApiKey:         *apiKey,
		ContextTimeout: *contextTimeout,
	}

	service := botservice.NewTgBotService(*tgBotToken, openaiConfig)
	service.Run()
}
