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
	CompletionTimeout    = 100 * time.Second
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

type qa struct {
	q string
	a string
	s string
}

type OpenAITalk struct {
	id           def.ConversationId
	bot          string
	greeting     qa
	messageQueue []qa
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
		greeting: qa{
			s: fmt.Sprintf(
				"You are a helpful assistant. Your name is %s.",
				cfg.BotName,
			),
		},
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
		if msg.s != "" {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: msg.s,
				},
			)
		}
		if msg.q != "" {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: msg.q,
				},
			)
		}
		if msg.a != "" {
			messages = append(messages,
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: msg.a,
				},
			)
		}
	}
	var resp openai.ChatCompletionResponse
	var err error
	retry := 3
	for retry > 0 {
		if func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), CompletionTimeout)
			defer cancel()
			resp, err = conv.client.CreateChatCompletion(
				ctx,
				openai.ChatCompletionRequest{
					Model:       conv.model,
					Messages:    messages,
					Temperature: 0.9,
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
		return "I apologize, but the OpenAI API is currently experiencing high traffic. Kindly try again at a later time."
	}

	answer := resp.Choices[0].Message.Content
	if answer != "" {
		conv.messageQueue[len(conv.messageQueue)-1].a = answer
	}
	return answer
}

func (conv *OpenAITalk) PrepareNewMessage(msg string) {
	totalTtoken := getTokenCount(msg) + getTokenCount(conv.greeting.s)
	newQueue := []qa{{q: msg}}

	now := time.Now()
	old := now.After(conv.lastMessage.Add(conv.contextAware))

	for i := len(conv.messageQueue) - 1; i > 0 && totalTtoken < MaxMessageQueueToken && !old; i-- {
		cnt := getTokenCount(conv.messageQueue[i].q) + getTokenCount(conv.messageQueue[i].a)
		if totalTtoken+cnt > MaxMessageQueueToken {
			break
		}
		newQueue = append(newQueue, conv.messageQueue[i])
		totalTtoken += cnt
	}
	newQueue = append(newQueue, conv.greeting)

	for i := 0; i < (len(newQueue) - 1 - i); i++ {
		m := newQueue[i]
		newQueue[i] = newQueue[len(newQueue)-1-i]
		newQueue[len(newQueue)-1-i] = m
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
