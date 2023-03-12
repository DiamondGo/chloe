/*
 * mastercoderk@gmail.com
 */
package ai

import (
	"chloe/def"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sashabaranov/go-openai"
)

const (
	MaxMessageQueueToken = 3000
)

// / singleton client
var lock = &sync.Mutex{}
var clients = make(map[string]*openai.Client)

func getOpenAIClient(apiKey string) *openai.Client {
	lock.Lock()
	defer lock.Unlock()

	client, exists := clients[apiKey]
	if exists {
		return client
	}

	client = openai.NewClient(apiKey)
	clients[apiKey] = client
	return client
}

type AIConfig struct {
	BotName string
	ApiKey  string
	Model   string
}

type OpenAITalk struct {
	id           def.ConversationId
	bot          string
	greeting     string
	messageQueue []string

	model  string
	client *openai.Client
}

var talkId int64 = 0

func NewTalk(cfg AIConfig) def.Conversation {
	talk := &OpenAITalk{
		id:       def.ConversationId(atomic.AddInt64(&talkId, 1)),
		bot:      cfg.BotName,
		greeting: fmt.Sprintf("Hello, I will call you %s in this conversation.", cfg.BotName),
		model:    cfg.Model,
		client:   getOpenAIClient(cfg.ApiKey),
	}
	return talk
}

func (conv *OpenAITalk) GetID() def.ConversationId {
	return conv.id
}

func (conv *OpenAITalk) Ask(q string) string {
	conv.PrepareNewMessage(q)

	var messages []openai.ChatCompletionMessage
	for _, msg := range conv.messageQueue {
		messages = append(messages,
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		)
	}
	resp, err := conv.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    conv.model,
			Messages: messages,
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return ""
	}
	return resp.Choices[0].Message.Content
}

func (conv *OpenAITalk) PrepareNewMessage(msg string) {
	totalTtoken := getTokenCount(conv.greeting) + getTokenCount(msg)

	newQueue := []string{msg}

	for i := len(conv.messageQueue) - 1; i > 0 && totalTtoken < MaxMessageQueueToken; i-- {
		cnt := getTokenCount(conv.messageQueue[i])
		if totalTtoken+cnt > MaxMessageQueueToken {
			break
		}
		newQueue = append(newQueue, conv.messageQueue[i])
		totalTtoken += cnt
	}
	newQueue = append(newQueue, conv.greeting)

	for i := 0; i < (len(newQueue) - 1 - i); i++ {
		msg := newQueue[i]
		newQueue[i] = newQueue[len(newQueue)-1-i]
		newQueue[len(newQueue)-1-i] = msg
	}

	conv.messageQueue = newQueue
}

func getTokenCount(msg string) int {
	words := strings.Fields(msg)
	return len(words)
}

type TalkFactory struct {
	talks  map[def.ChatID]def.Conversation
	config AIConfig
}

func NewTalkFactory(config AIConfig) def.ConversationFactory {
	return &TalkFactory{
		talks:  make(map[def.ChatID]def.Conversation),
		config: config,
	}
}

func (tf *TalkFactory) GetTalk(chatId def.ChatID) def.Conversation {
	talk, exists := tf.talks[chatId]
	if !exists {
		talk = NewTalk(tf.config)
		tf.talks[chatId] = talk
	}

	return talk
}
