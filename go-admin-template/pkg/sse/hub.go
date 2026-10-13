package sse

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// SSE 事件名约定
const (
	EventMessage = "message" // 新消息摘要
	EventUnread  = "unread"  // 最新未读数
)

// Event SSE 事件
type Event struct {
	Name string
	Data any
}

// Client 一个 SSE 连接
type Client struct {
	userID int
	ch     chan Event
}

// Hub 连接注册表（goroutine 安全）
type Hub struct {
	mu      sync.RWMutex
	clients map[int]map[*Client]struct{}
}

// DefaultHub 全局连接注册表
var DefaultHub = &Hub{clients: make(map[int]map[*Client]struct{})}

// Register 注册连接
func (h *Hub) Register(userID int) *Client {
	c := &Client{userID: userID, ch: make(chan Event, 8)}
	h.mu.Lock()
	if h.clients == nil {
		h.clients = make(map[int]map[*Client]struct{})
	}
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]struct{})
	}
	h.clients[userID][c] = struct{}{}
	h.mu.Unlock()
	logrus.Infof("sse client connected, userID=%d, online=%d", userID, h.OnlineCount())
	return c
}

// Unregister 注销连接并关闭事件通道
func (h *Hub) Unregister(userID int, c *Client) {
	h.mu.Lock()
	if set, ok := h.clients[userID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, userID)
		}
	}
	h.mu.Unlock()
	close(c.ch)
	logrus.Infof("sse client disconnected, userID=%d, online=%d", userID, h.OnlineCount())
}

// PushToUser 向指定用户的所有在线连接推送（缓冲满则丢弃，避免慢连接阻塞广播）
func (h *Hub) PushToUser(userID int, event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.ch <- Event{Name: event, Data: data}:
		default:
		}
	}
}

// PushToAll 向所有在线连接推送
func (h *Hub) PushToAll(event string, data any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.clients {
		for c := range set {
			select {
			case c.ch <- Event{Name: event, Data: data}:
			default:
			}
		}
	}
}

// OnlineCount 在线连接数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	n := 0
	for _, set := range h.clients {
		n += len(set)
	}
	return n
}
