/*
 * mastercoderk@gmail.com
 */

package def

/// IM interface

type ChatID string
type UserID string
type MessageID string
type CleanFunc func()

func (uid UserID) String() string {
	return string(uid)
}

func (cid ChatID) String() string {
	return string(cid)
}

func (mid MessageID) String() string {
	return string(mid)
}

type Chat interface {
	GetID() ChatID
	GetMemberCount() int
	SendMessage(string)
	ReplyMessage(string, MessageID)
	QuoteMessage(message string, replyTo MessageID, quote string)
	ReplyImage(string, MessageID)
	ReplyVoice(aud string, to MessageID)
	GetSelf() User
}

type User interface {
	GetID() UserID
	GetFirstName() string
	GetUserName() string
}

type Message interface {
	GetID() MessageID
	GetUser() User
	GetChat() Chat
	GetText() string
	GetVoice() (string, CleanFunc)
}

type MessageBot interface {
	GetMessages() <-chan Message
}

type Debuggable interface {
	SetDebug(bool)
}

/// AI interface

type ConversationId int64

type Conversation interface {
	GetID() ConversationId
	Ask(string) string
}

type ConversationFactory interface {
	GetTalk(ChatID) Conversation
}

type SpeechToText interface {
	Convert(voiceFile string) (string, error)
}

type TextToSpeech interface {
	Convert(text string) (string, CleanFunc, error)
}

type ImageGenerator interface {
	Generate(desc, size string) (string, CleanFunc, error)
}

/// for service

type BotService interface {
	Run()
}
