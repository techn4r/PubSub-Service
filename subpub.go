package main

import (
	"context"
	"fmt"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subPubImpl struct {
	mu      sync.RWMutex
	subs    map[string][]MessageHandler
	closing bool
}

func NewSubPub() SubPub {
	return &subPubImpl{
		subs: make(map[string][]MessageHandler),
	}
}

func (s *subPubImpl) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return nil, fmt.Errorf("subpub is closing")
	}

	s.subs[subject] = append(s.subs[subject], cb)

	return &subscriptionImpl{
		subject: subject,
		cb:      cb,
		parent:  s,
	}, nil
}

func (s *subPubImpl) Publish(subject string, msg interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	handlers, ok := s.subs[subject]
	if !ok {
		return nil
	}

	for _, handler := range handlers {
		handler(msg)
	}

	return nil
}

func (s *subPubImpl) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closing = true
	s.subs = make(map[string][]MessageHandler)
	return nil
}

type subscriptionImpl struct {
	subject string
	cb      MessageHandler
	parent  *subPubImpl
}

func (s *subscriptionImpl) Unsubscribe() {
	s.parent.mu.Lock()
	defer s.parent.mu.Unlock()

	handlers := s.parent.subs[s.subject]
	for i, handler := range handlers {
		if fmt.Sprintf("%p", handler) == fmt.Sprintf("%p", s.cb) {
			s.parent.subs[s.subject] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	if len(s.parent.subs[s.subject]) == 0 {
		delete(s.parent.subs, s.subject)
	}
}
