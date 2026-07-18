// Package app 承载用例层：在 repository 之上编排业务规则，对外暴露 Usecase 接口。
package app

import (
	"sync"

	"github.com/itgrape/pcc/server/internal/model"
)

// EventType 事件类型。
type EventType string

const (
	EventConfigUpdated EventType = "config.updated"
	EventConfigDeleted EventType = "config.deleted"
)

// ConfigEvent 配置变更事件。
type ConfigEvent struct {
	Type         EventType
	NamespaceID  uint64
	Namespace    string
	ConfigID     uint64
	ConfigName   string
	Format       model.ConfigFormat
	Content      string
	Version      uint
}

// Subscriber 事件订阅者（通常是 WebSocket 管理器）。
type Subscriber interface {
	OnConfigEvent(ConfigEvent)
}

// EventBus 轻量级内存事件总线。
type EventBus struct {
	mu   sync.RWMutex
	subs []Subscriber
}

// NewEventBus 创建事件总线。
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe 注册订阅者。
func (b *EventBus) Subscribe(s Subscriber) {
	b.mu.Lock()
	b.subs = append(b.subs, s)
	b.mu.Unlock()
}

// Unsubscribe 移除订阅者。
func (b *EventBus) Unsubscribe(s Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, sub := range b.subs {
		if sub == s {
			b.subs = append(b.subs[:i], b.subs[i+1:]...)
			return
		}
	}
}

// Publish 发布事件到所有订阅者。
func (b *EventBus) Publish(e ConfigEvent) {
	b.mu.RLock()
	subs := make([]Subscriber, len(b.subs))
	copy(subs, b.subs)
	b.mu.RUnlock()
	for _, s := range subs {
		s.OnConfigEvent(e)
	}
}
