/*
 * mastercoderk@gmail.com
 */
package def

/// IM interface

type ChatID int64
type UserID int64
type MessageID int

type Chat interface {
	GetID() ChatID
	GetMemberCount() int
	SendMessage(string)
	ReplyMessage(string, MessageID)
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

/// for service

type BotService interface {
	Run()
}
