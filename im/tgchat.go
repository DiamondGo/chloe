/*
 * mastercoderk@gmail.com
 */
package im

import (
	"chloe/def"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type tgBotCache struct {
	chats map[def.ChatID]def.Chat
	users map[def.ChatID]map[def.UserID]def.User
}

type TelegramBot struct {
	msgQueue chan def.Message
	api      *tgbotapi.BotAPI
	cache    tgBotCache
}

func NewTelegramBot(token string) (def.MessageBot, error) {
	bot := &TelegramBot{
		msgQueue: make(chan def.Message, 100),
		cache: tgBotCache{
			chats: make(map[def.ChatID]def.Chat),
			users: make(map[def.ChatID]map[def.UserID]def.User),
		},
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to initialize telegram bot")
		return nil, err
	}
	bot.api = api

	go bot.messageLoop()

	return bot, nil
}

func (bot *TelegramBot) GetMessages() <-chan def.Message {
	return bot.msgQueue
}

func (bot *TelegramBot) SetDebug(debug bool) {
	bot.api.Debug = debug
}

func (bot *TelegramBot) messageLoop() {
	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 120

	updates := bot.api.GetUpdatesChan(cfg)

	for update := range updates {
		if update.Message != nil { // If we got a message
			m := &tgMessage{
				id:     def.MessageID(update.Message.MessageID),
				userId: def.UserID(update.Message.From.ID),
				chatId: def.ChatID(update.Message.Chat.ID),
				text:   update.Message.Text,
				bot:    bot,
			}

			bot.msgQueue <- m
		}
	}
}

func (bot *TelegramBot) lookupChat(id def.ChatID) def.Chat {
	if chat, exists := bot.cache.chats[id]; exists {
		return chat
	}

	var chat def.Chat

	count, _ := bot.api.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: int64(id),
		},
	})
	chat = &tgChat{
		id:          id,
		memberCount: count,
		bot:         bot,
	}

	bot.cache.chats[id] = chat

	return chat
}

func (bot *TelegramBot) lookupUser(uid def.UserID, cid def.ChatID) def.User {
	if user, exists := bot.cache.users[cid][uid]; exists {
		return user
	}

	var user def.User
	chatMembersConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			UserID: int64(uid),
			ChatID: int64(cid),
		},
	}
	chatMember, err := bot.api.GetChatMember(chatMembersConfig)
	if err != nil {
		log.Fatalf("failed to get user %v", uid)
		return nil
	}
	user = &tgUser{
		id:        uid,
		firstName: chatMember.User.FirstName,
		userName:  chatMember.User.UserName,
		chatId:    cid,
	}

	chat, exists := bot.cache.users[cid]
	if !exists {
		chat = make(map[def.UserID]def.User)
		bot.cache.users[cid] = chat
	}
	chat[uid] = user

	return user
}

type tgMessage struct {
	id     def.MessageID
	userId def.UserID
	chatId def.ChatID
	text   string

	bot *TelegramBot
}

func (m *tgMessage) GetID() def.MessageID {
	return m.id
}

func (m *tgMessage) GetUser() def.User {
	return m.bot.lookupUser(m.userId, m.chatId)
}

func (m *tgMessage) GetChat() def.Chat {
	return m.bot.lookupChat(m.chatId)
}

func (m *tgMessage) GetText() string {
	return m.text
}

type tgChat struct {
	id          def.ChatID
	memberCount int

	bot *TelegramBot
}

func (c *tgChat) GetID() def.ChatID {
	return c.id
}

func (c *tgChat) GetMemberCount() int {
	return c.memberCount
}

func (c *tgChat) SendMessage(m string) {
	// TODO
}

func (c *tgChat) ReplyMessage(m string, to def.MessageID) {

	mksafe := escapeSafeForMarkdown(m)
	msg := tgbotapi.NewMessage(int64(c.id), mksafe)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyToMessageID = int(to)

	_, err := c.bot.api.Send(msg)
	if err != nil {
		log.Printf("error: %#v in sending message: %#v", err, msg)
		fallbackMsg := tgbotapi.NewMessage(int64(c.id), m)
		fallbackMsg.ParseMode = ""
		fallbackMsg.ReplyToMessageID = int(to)
		_, err := c.bot.api.Send(fallbackMsg)
		if err != nil {
			log.Printf("error: %#v in retry sending message: %#v", err, fallbackMsg)
		}
	}
}

func (c *tgChat) GetSelf() def.User {
	return &tgUser{
		id:        def.UserID(c.bot.api.Self.ID),
		firstName: c.bot.api.Self.FirstName,
		userName:  c.bot.api.Self.UserName,
		chatId:    c.id,
	}
}

type tgUser struct {
	id        def.UserID
	firstName string
	userName  string

	chatId def.ChatID
}

func (u *tgUser) GetID() def.UserID {
	return u.id
}

func (u *tgUser) GetFirstName() string {
	return u.firstName
}

func (u *tgUser) GetUserName() string {
	return u.userName
}

func escapeSafeForMarkdown(s string) string {
	s = strings.ReplaceAll(s, "!", `\!`)
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	s = strings.ReplaceAll(s, "-", `\-`)
	s = strings.ReplaceAll(s, "+", `\+`)
	s = strings.ReplaceAll(s, ">", `\>`)
	s = strings.ReplaceAll(s, "<", `\<`)

	return s
}
