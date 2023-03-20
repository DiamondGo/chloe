/*
 * mastercoderk@gmail.com
 */
package ai

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"chloe/def"
	"context"

	log "github.com/jeanphorn/log4go"
	"github.com/sashabaranov/go-openai"
)

const (
	MaxMessageQueueToken = 3000
	ContextAwareTime     = time.Minute
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
	BotName        string
	ApiKey         string
	Model          string
	ContextTimeout int
}

type OpenAITalk struct {
	id           def.ConversationId
	bot          string
	greeting     string
	messageQueue []string
	lastMessage  time.Time
	contextAware time.Duration

	model  string
	client *openai.Client
}

var talkId int64 = 0

func NewTalk(cfg AIConfig) def.Conversation {
	ctxTimeout := time.Duration(cfg.ContextTimeout) * time.Second
	talk := &OpenAITalk{
		id:  def.ConversationId(atomic.AddInt64(&talkId, 1)),
		bot: cfg.BotName,
		greeting: fmt.Sprintf(
			"Hello, can I call you Chloe in our following coversation?",
			cfg.BotName,
		),
		model:        cfg.Model,
		client:       getOpenAIClient(cfg.ApiKey),
		lastMessage:  time.Time{},
		contextAware: ctxTimeout,
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
	var resp openai.ChatCompletionResponse
	var err error
	retry := 2
	for retry > 0 {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			resp, err = conv.client.CreateChatCompletion(
				ctx,
				openai.ChatCompletionRequest{
					Model:    conv.model,
					Messages: messages,
				},
			)
			conv.lastMessage = time.Now()

			if err != nil {
				log.Info("ChatCompletion error: %v\n", err)
				return false
			}

			return true
		}() {
			break
		}
		retry--
	}

	if err != nil {
		log.Info("failed to get response from openai.")
		return ""
	}

	return resp.Choices[0].Message.Content
}

func (conv *OpenAITalk) PrepareNewMessage(msg string) {
	totalTtoken := getTokenCount(conv.greeting) + getTokenCount(msg)

	newQueue := []string{msg}

	now := time.Now()
	old := now.After(conv.lastMessage.Add(conv.contextAware))

	for i := len(conv.messageQueue) - 1; i > 0 && totalTtoken < MaxMessageQueueToken && !old; i-- {
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
