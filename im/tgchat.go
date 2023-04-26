/*
 * mastercoderk@gmail.com
 */

package im

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"chloe/def"
	"chloe/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/jeanphorn/log4go"
)

const (
	// prefix for Telegram IDs
	preTG = "tg-"
)

type tgBotCache struct {
	guard sync.Mutex
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
		log.Error("failed to initialize telegram bot")
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
			var m def.Message
			if update.Message.Voice != nil {
				// process voice
				fd := update.Message.Voice.FileID
				link, err := bot.api.GetFileDirectURL(fd)
				if err != nil {
					log.Warn("failed to get voice download link")
				}
				log.Debug("voice file link %s", link)
				voiceFile, cleaner := util.DownloadTempFile(link)
				m = &tgMessage{
					id: def.MessageID(
						preTG + strconv.FormatInt(int64(update.Message.MessageID), 10),
					),
					userId:     def.UserID(preTG + strconv.FormatInt(update.Message.From.ID, 10)),
					chatId:     def.ChatID(preTG + strconv.FormatInt(update.Message.Chat.ID, 10)),
					bot:        bot,
					audioFile:  voiceFile,
					audioClean: cleaner,
				}
			} else if update.Message.Text != "" {
				m = &tgMessage{
					id:     def.MessageID(preTG + strconv.FormatInt(int64(update.Message.MessageID), 10)),
					userId: def.UserID(preTG + strconv.FormatInt(update.Message.From.ID, 10)),
					chatId: def.ChatID(preTG + strconv.FormatInt(update.Message.Chat.ID, 10)),
					bot:    bot,
					text:   update.Message.Text,
				}
			}

			bot.msgQueue <- m
		}
	}
}

func (bot *TelegramBot) lookupChat(id def.ChatID) def.Chat {
	var chat def.Chat
	if chat = bot.cache.getChatById(id); chat != nil {
		return chat
	}

	count, _ := bot.api.GetChatMembersCount(tgbotapi.ChatMemberCountConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: bot.getInt64ChatId(id),
		},
	})
	chat = &tgChat{
		id:          id,
		memberCount: count,
		bot:         bot,
	}

	bot.cache.cacheChatById(id, chat)

	return chat
}

func (bot *TelegramBot) lookupUser(uid def.UserID, cid def.ChatID) def.User {
	if user := bot.cache.getChatUser(cid, uid); user != nil {
		return user
	}

	var user def.User
	chatMembersConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			UserID: bot.getInt64UserId(uid),
			ChatID: bot.getInt64ChatId(cid),
		},
	}
	chatMember, err := bot.api.GetChatMember(chatMembersConfig)
	if err != nil {
		log.Error("failed to get user %v", uid)
		return nil
	}
	user = &tgUser{
		id:        uid,
		firstName: chatMember.User.FirstName,
		userName:  chatMember.User.UserName,
		chatId:    cid,
	}

	bot.cache.cacheChatUser(cid, uid, user)

	return user
}

type tgMessage struct {
	id         def.MessageID
	userId     def.UserID
	chatId     def.ChatID
	text       string
	audioFile  string
	audioClean def.CleanFunc

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

func (m *tgMessage) GetVoice() (string, def.CleanFunc) {
	return m.audioFile, m.audioClean
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
	msg := tgbotapi.NewMessage(c.bot.getInt64ChatId(c.id), mksafe)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyToMessageID = c.bot.getIntMessageId(to)

	_, err := c.bot.api.Send(msg)
	if err != nil {
		log.Info("error: %#v in sending message: %#v", err, msg)
		fallbackMsg := tgbotapi.NewMessage(c.bot.getInt64ChatId(c.id), m)
		fallbackMsg.ParseMode = ""
		fallbackMsg.ReplyToMessageID = c.bot.getIntMessageId(to)
		_, err := c.bot.api.Send(fallbackMsg)
		if err != nil {
			log.Info("error: %#v in retry sending message: %#v", err, fallbackMsg)
		}
	}
}

func (c *tgChat) QuoteMessage(m string, to def.MessageID, quote string) {
	mksafe := "_*" + escapeSafeForMarkdown(quote) + "*_"
	mksafe += "  \n"
	mksafe += "  \n"
	mksafe += escapeSafeForMarkdown(m)

	msg := tgbotapi.NewMessage(c.bot.getInt64ChatId(c.id), mksafe)
	msg.ParseMode = "MarkdownV2"
	msg.ReplyToMessageID = c.bot.getIntMessageId(to)

	_, err := c.bot.api.Send(msg)
	if err != nil {
		log.Info("error: %#v in sending message: %#v", err, mksafe)
	}
}

func (c *tgChat) ReplyImage(img string, to def.MessageID) {
	f, err := os.Open(img)
	if err != nil {
		log.Error("open image file %s failed, %v", img, err)
		return
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		log.Error("stat image file %s failed, %v", img, err)
		return
	}
	fileSize := fileInfo.Size()

	buffer := make([]byte, fileSize)
	_, err = f.Read(buffer)
	if err != nil {
		log.Error("read image file %s failed, %v", img, err)
		return
	}

	requestFileData := &tgbotapi.FileBytes{
		Name:  filepath.Base(f.Name()),
		Bytes: buffer,
	}

	photoMsg := tgbotapi.NewPhoto(c.bot.getInt64ChatId(c.id), requestFileData)
	photoMsg.ReplyToMessageID = c.bot.getIntMessageId(to)
	if _, err = c.bot.api.Send(photoMsg); err != nil {
		log.Error("failed to send image to user")
	}
}

func (c *tgChat) ReplyVoice(aud string, to def.MessageID) {
	f, err := os.Open(aud)
	if err != nil {
		log.Error("open audio file %s failed, %v", aud, err)
		return
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		log.Error("stat audio file %s failed, %v", aud, err)
		return
	}
	fileSize := fileInfo.Size()

	buffer := make([]byte, fileSize)
	_, err = f.Read(buffer)
	if err != nil {
		log.Error("read audio file %s failed, %v", aud, err)
		return
	}

	requestFileData := &tgbotapi.FileBytes{
		Name:  filepath.Base(f.Name()),
		Bytes: buffer,
	}

	/*
		var media tgbotapi.MediaGroupConfig
		vm := tgbotapi.NewInputMediaAudio(requestFileData)
		media.Media = append(media.Media, vm)
		voiceMsg := tgbotapi.NewMediaGroup(int64(c.id), media.Media)
	*/
	voiceMsg := tgbotapi.NewAudio(c.bot.getInt64ChatId(c.id), requestFileData)
	voiceMsg.ReplyToMessageID = c.bot.getIntMessageId(to)
	if _, err = c.bot.api.Send(voiceMsg); err != nil {
		log.Error("failed to send image to user %v", err)
		return
	}
	log.Debug("send size %d voice file %s", fileSize, aud)

}

func (c *tgChat) GetSelf() def.User {
	return &tgUser{
		id:        def.UserID(preTG + strconv.FormatInt(c.bot.api.Self.ID, 10)),
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
	s = strings.ReplaceAll(s, "=", `\=`)

	return s
}

func (c *tgBotCache) getChatById(id def.ChatID) def.Chat {
	c.guard.Lock()
	defer c.guard.Unlock()

	if chat, exists := c.chats[id]; exists {
		return chat
	}
	return nil
}

func (c *tgBotCache) cacheChatById(id def.ChatID, chat def.Chat) {
	c.guard.Lock()
	defer c.guard.Unlock()

	c.chats[id] = chat
}

func (c *tgBotCache) getChatUser(cid def.ChatID, uid def.UserID) def.User {
	c.guard.Lock()
	defer c.guard.Unlock()

	if user, exists := c.users[cid][uid]; exists {
		return user
	}
	return nil
}

func (c *tgBotCache) cacheChatUser(cid def.ChatID, uid def.UserID, user def.User) {
	c.guard.Lock()
	defer c.guard.Unlock()

	chat, exists := c.users[cid]
	if !exists {
		chat = make(map[def.UserID]def.User)
		c.users[cid] = chat
	}
	chat[uid] = user
}

func (bot *TelegramBot) getInt64ChatId(cid def.ChatID) int64 {
	id := string(cid)
	id = id[len(preTG):]
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func (bot *TelegramBot) getInt64UserId(uid def.UserID) int64 {
	id := string(uid)
	id = id[len(preTG):]
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func (bot *TelegramBot) getIntMessageId(mid def.MessageID) int {
	id := string(mid)
	id = id[len(preTG):]
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0
	}
	return int(n)
}
