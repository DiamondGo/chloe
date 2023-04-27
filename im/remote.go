/*
 * mastercoderk@gmail.com
 */

package im

import (
	"chloe/def"
	"io"
	"net"
	"strings"

	psg "chloe/proto/service/go"

	log "github.com/jeanphorn/log4go"
	"google.golang.org/grpc"
)

const (
	// remote message, for M$ Teams or else
	preRM = "rm-"
)

func NewRemoteChatBot(port string) (def.MessageBot, error) {
	bot := &remoteBot{
		msgQueue:  make(chan def.Message, 100),
		outgoingQ: make(chan *psg.Message, 100),
		cache:     newChatCache(),
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Error("fail to start grpc server on port %s", port, err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	psg.RegisterChattingServer(grpcServer, newRemoteChatServer(bot))
	go grpcServer.Serve(lis)

	return bot, nil
}

type remoteChatServer struct {
	psg.UnimplementedChattingServer // wtf is this?

	bot *remoteBot
}

func newRemoteChatServer(bot *remoteBot) psg.ChattingServer {
	return &remoteChatServer{
		bot: bot,
	}
}

func (s *remoteChatServer) ChatStream(stream psg.Chatting_ChatStreamServer) error {
	/*
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			reply := "re: " + msg.Text

			msgReply := &psg.Message{
				ReplyToId: msg.Id,
				Text:      reply,
				Chat:      msg.Chat,
			}

			stream.Send(msgReply)
		}
	*/
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				// TODO log
				break
			}

			chat := msg.Chat
			chatId := chat.Id

			user := msg.Sender

			rMsg := &remoteMessage{
				bot:  s.bot,
				id:   msg.Id,
				text: msg.Text,
				chat: &remoteChat{
					bot: s.bot,
					id:  chatId,
				},
				from: &remoteUser{
					id:       user.Id,
					username: user.UserName,
				},
			}

			s.bot.msgQueue <- rMsg
		}
	}()

	for msg := range s.bot.outgoingQ {
		stream.Send(msg)
	}

	return nil
}

type remoteBot struct {
	msgQueue  chan def.Message
	outgoingQ chan *psg.Message
	cache     *chatCache
}

func (bot *remoteBot) GetMessages() <-chan def.Message {
	return bot.msgQueue
}

type remoteMessage struct {
	bot  *remoteBot
	id   string
	text string
	chat *remoteChat
	from *remoteUser
}

type remoteChat struct {
	bot *remoteBot
	id  string
}

type remoteUser struct {
	id       string
	username string
}

func (m *remoteMessage) GetChat() def.Chat {
	return m.chat
}

func (m *remoteMessage) GetID() def.MessageID {
	return def.MessageID(preRM + m.id)
}

func (m *remoteMessage) GetText() string {
	return m.text
}

func (m *remoteMessage) GetUser() def.User {
	return m.from
}

func (m *remoteMessage) GetVoice() (string, def.CleanFunc) {
	// TODO
	return "", func() {}
}

// User
func (u *remoteUser) GetID() def.UserID {
	return def.UserID(preRM + u.id)
}

func (u *remoteUser) GetFirstName() string {
	// TODO
	return ""
}

func (u *remoteUser) GetUserName() string {
	return u.username
}

// Chat
func (c *remoteChat) GetID() def.ChatID {
	return def.ChatID(preRM + c.id)
}

func (c *remoteChat) GetMemberCount() int {
	// TODO: may support group chat later
	return 2
}

func (c *remoteChat) SendMessage(m string) {
	// TODO
}

func (c *remoteChat) GetSelf() def.User {
	// TODO: change later, need rpc support
	return &remoteUser{
		id:       preRM + "self",
		username: "Chloe",
	}
}

func (c *remoteChat) QuoteMessage(m string, to def.MessageID, quote string) {
	// TODO
	c.ReplyMessage(m, to)
}

func (c *remoteChat) ReplyMessage(m string, to def.MessageID) {
	msgReply := &psg.Message{
		ReplyToId: c.stripId(to.String()),
		Text:      m,
		Chat: &psg.Chat{
			Id: c.stripId(c.id),
		},
	}
	c.bot.outgoingQ <- msgReply
}

func (c *remoteChat) ReplyImage(img string, to def.MessageID) {
	// TODO
	panic("unimplemented")
}

func (c *remoteChat) ReplyVoice(aud string, to def.MessageID) {
	// TODO
	panic("unimplemented")
}

func (c *remoteChat) stripId(id string) string {
	if strings.HasPrefix(id, preRM) {
		return id[len(preRM):]
	}
	return id
}
