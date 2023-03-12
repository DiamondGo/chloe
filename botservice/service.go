/*
 * mastercoderk@gmail.com
 */package botservice

import (
	"log"
	"strings"

	"chloe/ai"
	"chloe/def"
	"chloe/im"
)

var puncs = []string{",", ".", "，", "。", "!", "?", "！", "？"}

type BotTalkService struct {
	bot      def.MessageBot
	talkFact def.ConversationFactory
	config   ai.AIConfig
}

func NewTgBotService(tgbotToken string, aicfg ai.AIConfig) def.BotService {
	bot, err := im.NewTelegramBot("6062850243:AAG4Fn7i-NXI8o_VuWBdazPTlezV6ql4eVs")
	if err != nil {
		log.Panic(err)
	}
	return &BotTalkService{
		bot:      bot,
		talkFact: ai.NewTalkFactory(aicfg),
		config:   aicfg,
	}
}

func (s *BotTalkService) Run() {
	for m := range s.bot.GetMessages() {
		chat := m.GetChat()
		memberCnt := chat.GetMemberCount()
		text := m.GetText()
		botUsername := chat.GetSelf().GetUserName()
		if memberCnt > 2 && !s.isMentioned(text, botUsername) {
			continue
		}

		talk := s.talkFact.GetTalk(chat.GetID())
		answer := talk.Ask(text)

		chat.ReplyMessage(answer, m.GetID())
	}

}

func (s *BotTalkService) isMentioned(text, botUsername string) bool {
	tokens := strings.Fields(text)
	if ssContain(tokens, "@"+botUsername, false) {
		return true
	}

	tokens = spliteByPunctuation(tokens)
	var head []string
	if len(tokens) >= 3 {
		head = tokens[:3]
	} else {
		head = tokens
	}

	if ssContain(head, s.config.BotName, true) {
		return true
	}

	return false
}

func ssContain(ss []string, target string, ignoreCase bool) bool {
	for _, s := range ss {
		if target == removePunctuation(s) ||
			(ignoreCase && strings.EqualFold(target, removePunctuation(s))) {
			return true
		}
	}
	return false
}

func removePunctuation(s string) string {
	for _, p := range puncs {
		s = strings.Replace(s, p, "", -1)
	}

	return s
}

func spliteByPunctuation(ss []string) []string {
	for _, p := range puncs {
		newSs := []string{}
		for _, s := range ss {
			sp := strings.Split(s, p)
			newSs = append(newSs, sp...)
		}
		ss = newSs
	}

	return ss
}
