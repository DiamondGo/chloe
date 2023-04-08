/*
 * mastercoderk@gmail.com
 */

package botservice

import (
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"chloe/ai"
	"chloe/def"
	"chloe/im"
	"chloe/util"

	"github.com/DiamondGo/gohelper"
	log "github.com/jeanphorn/log4go"
)

var puncs = []string{",", ".", "，", "。", "!", "?", "！", "？"}

type BotTalkService struct {
	bot            def.MessageBot
	talkFact       def.ConversationFactory
	speechToText   def.SpeechToText
	textToSpeech   def.TextToSpeech
	imageGenerator def.ImageGenerator
	config         ai.AIConfig
	loop           bool
}

func NewTgBotService(tgbotToken string, aicfg ai.AIConfig) def.BotService {
	bot, err := im.NewTelegramBot(tgbotToken)
	if err != nil {
		log.Error("failed to start telegram bot %v", err)
	}
	return &BotTalkService{
		bot:            bot,
		talkFact:       ai.NewTalkFactory(aicfg),
		speechToText:   ai.NewSpeech2Text(aicfg.ApiKey),
		textToSpeech:   ai.NewPyServiceTTS(),
		imageGenerator: ai.NewImageGenerator(aicfg.ApiKey),
		config:         aicfg,
		loop:           true,
	}
}

func (s *BotTalkService) Run() {
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)
	go func() {
		for {
			sig := <-sigchnl
			s.handleSignal(sig)
		}
	}()

	pool := gohelper.NewTaskPool[def.UserID](3, 1)
	for m := range s.bot.GetMessages() {
		var uid def.UserID
		if m == nil || m.GetUser() == nil || m.GetUser().GetID() == uid || m.GetChat() == nil {
			log.Warn("received empty message, skip it")
			if m == nil {
				log.Debug("message is nil")
				continue
			}
			log.Debug("message user is %v", m.GetUser())
			log.Debug("message chat is %v", m.GetChat())
			continue
		}

		user := m.GetUser()
		uid = user.GetID()
		chat := m.GetChat()
		voice, voiceCleaner := m.GetVoice()
		msgText := m.GetText()
		msgID := m.GetID()

		task := func() {
			defer func() { _ = recover() }()
			memberCnt := chat.GetMemberCount()
			botUsername := chat.GetSelf().GetUserName()

			var text string
			var err error

			if voice != "" {
				defer voiceCleaner()
				// get text from voice
				var mp3 string
				var cleaner def.CleanFunc
				if strings.EqualFold(filepath.Ext(voice), ".oga") {
					mp3, cleaner = util.ConvertToMp3(voice)
					defer cleaner()
				}
				text, err = s.speechToText.Convert(mp3)
				if err != nil {
					log.Warn("speech to text failed, %v", err)
				}
			} else if msgText != "" {
				text = msgText
			}

			if size, desc := s.isDrawCommand(text); size != "" {
				// draw image
				log.Debug(
					"received image request from %s, id %d: %s",
					user.GetUserName(),
					uid,
					text,
				)
				img, cleaner, err := s.imageGenerator.Generate(desc, size)
				if err != nil {
					chat.ReplyMessage(err.Error(), msgID)
					return
				}
				defer cleaner()
				chat.ReplyImage(img, msgID)
			} else {
				if memberCnt > 2 && !s.isMentioned(text, botUsername) {
					return
				}

				log.Info("received question from %s, id %d: %s", user.GetUserName(), uid, text)
				talk := s.talkFact.GetTalk(chat.GetID())
				answer := talk.Ask(text)

				if voice == "" {
					chat.ReplyMessage(answer, msgID)
				} else {
					chat.QuoteMessage(answer, msgID, "Transcription:\n"+text)
					if vf, cleaner, err := s.textToSpeech.Convert(answer); err != nil {
						log.Error(`convert text "%s" to speech failed, %v`, text, err)
					} else {
						defer cleaner()
						chat.ReplyVoice(vf, msgID)
						log.Info("voice replied to %s", user.GetUserName())
					}
				}
				log.Info("replied to %s", user.GetUserName())
			}
		}
		pool.Run(uid, task)
		if !s.loop {
			log.Info("stop running loop")
			time.Sleep(time.Duration(time.Second * 3))
			break
		}
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

func (s *BotTalkService) handleSignal(signal os.Signal) {
	switch signal {
	case syscall.SIGQUIT:
		s.loop = false
		log.Info("signal SIGQUIT received")
	default:
	}
}

func (s *BotTalkService) isDrawCommand(text string) (string, string) {
	pat := regexp.MustCompile(`^/(?P<drawcmd>(draw|drawbig|drawsmall))\s+(?P<desc>.*)`)
	match := pat.FindStringSubmatch(text)

	result := make(map[string]string)
	for i, name := range pat.SubexpNames() {
		if i != 0 && name != "" && len(match) > i {
			result[name] = match[i]
		}
	}

	cmd, _ := result["drawcmd"]
	desc, _ := result["desc"]
	switch cmd {
	case "draw":
		return "m", desc
	case "drawbig":
		return "b", desc
	case "drawsmall":
		return "s", desc
	default:
		return "", ""
	}
}
