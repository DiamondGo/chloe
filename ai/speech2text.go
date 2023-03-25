/*
 * mastercoderk@gmail.com
 */
package ai

import (
	"chloe/def"
	"context"
	"time"

	log "github.com/jeanphorn/log4go"
	"github.com/sashabaranov/go-openai"
)

type whisper struct {
	client *openai.Client
}

func NewSpeech2Text(apiKey string) def.SpeechToText {
	client := getOpenAIClient(apiKey)
	return &whisper{
		client: client,
	}
}

func (w *whisper) Convert(voiceFile string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*45)
	defer cancel()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: voiceFile,
	}

	resp, err := w.client.CreateTranscription(ctx, req)
	if err != nil {
		log.Error("failed to get speech transcripted, %v", err)
		return "", err
	}

	return resp.Text, nil
}
