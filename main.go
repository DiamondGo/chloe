package main

import (
	"flag"

	"chloe/ai"
	"chloe/botservice"

	"github.com/sashabaranov/go-openai"
)

func main() {
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
