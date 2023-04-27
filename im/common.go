/*
 * mastercoderk@gmail.com
 */

package im

import (
	"chloe/def"
	"sync"
)

type chatCache struct {
	guard sync.Mutex
	chats map[def.ChatID]def.Chat
	users map[def.ChatID]map[def.UserID]def.User
}

func (c *chatCache) getChat(id def.ChatID) def.Chat {
	c.guard.Lock()
	defer c.guard.Unlock()

	if chat, exists := c.chats[id]; exists {
		return chat
	}
	return nil
}

func (c *chatCache) cacheChat(id def.ChatID, chat def.Chat) {
	c.guard.Lock()
	defer c.guard.Unlock()

	c.chats[id] = chat
}

func (c *chatCache) getChatUser(cid def.ChatID, uid def.UserID) def.User {
	c.guard.Lock()
	defer c.guard.Unlock()

	if user, exists := c.users[cid][uid]; exists {
		return user
	}
	return nil
}

func (c *chatCache) cacheChatUser(cid def.ChatID, uid def.UserID, user def.User) {
	c.guard.Lock()
	defer c.guard.Unlock()

	chat, exists := c.users[cid]
	if !exists {
		chat = make(map[def.UserID]def.User)
		c.users[cid] = chat
	}
	chat[uid] = user
}

func newChatCache() *chatCache {
	return &chatCache{
		chats: make(map[def.ChatID]def.Chat),
		users: make(map[def.ChatID]map[def.UserID]def.User),
	}
}
