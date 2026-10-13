package sse

import (
	"context"
	"encoding/json"

	"go-admin-template/pkg/redis"

	"github.com/sirupsen/logrus"
)

// Channel Redis Pub/Sub 频道
const Channel = "message:new"

// MessagePayload 新消息广播负载
type MessagePayload struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	PublishTime string `json:"publish_time"`
}

// PublishNewMessage 广播新消息（写库成功后调用）
// 各实例订阅该频道后推送给本机在线连接，实现多实例扇出
func PublishNewMessage(payload MessagePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return redis.Client.Publish(context.Background(), Channel, data).Err()
}

// StartSubscriber 启动 Redis 订阅协程（进程启动时调用一次）
func StartSubscriber() {
	go func() {
		sub := redis.Client.Subscribe(context.Background(), Channel)
		defer func() {
			_ = sub.Close()
		}()
		logrus.Info("sse subscriber started, channel=", Channel)
		for msg := range sub.Channel() {
			var payload MessagePayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				logrus.Warnf("sse subscriber invalid payload: %v", err)
				continue
			}
			DefaultHub.PushToAll(EventMessage, payload)
		}
		logrus.Warn("sse subscriber channel closed")
	}()
}
