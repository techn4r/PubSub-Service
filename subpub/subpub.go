package subpub

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("subpub is closed")

// MessageHandler вызывается на каждое полученное сообщение.
type MessageHandler func(msg interface{})

// Subscribtion представляет собой возможность отписаться.
type Subscribtion interface {
	Unsubscribe()
}

// SubPub — основной интерфейс.
type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscribtion, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

type subPub struct {
	mu        sync.RWMutex
	subs      map[string]map[*subscription]struct{}
	closed    bool
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type subscription struct {
	mu      sync.Mutex
	cond    *sync.Cond
	queue   []interface{}
	cb      MessageHandler
	closed  bool
	sp      *subPub
	subject string
}

// NewSubPub создает новый инстанс SubPub.
func NewSubPub() SubPub {
	return &subPub{
		subs: make(map[string]map[*subscription]struct{}),
	}
}

// Subscribe регистрирует callback на subject и запускает горутину-обработчик.
func (sp *subPub) Subscribe(subject string, cb MessageHandler) (Subscribtion, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.closed {
		return nil, ErrClosed
	}

	s := &subscription{
		cb:      cb,
		queue:   make([]interface{}, 0),
		sp:      sp,
		subject: subject,
	}
	s.cond = sync.NewCond(&s.mu)

	if sp.subs[subject] == nil {
		sp.subs[subject] = make(map[*subscription]struct{})
	}
	sp.subs[subject][s] = struct{}{}

	// Увеличиваем счетчик потоков и запускаем цикл обработки
	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		s.loop()
	}()
	return s, nil
}

// Publish рассылает msg всем подписчикам по заданному subject.
func (sp *subPub) Publish(subject string, msg interface{}) error {
	sp.mu.RLock()
	if sp.closed {
		sp.mu.RUnlock()
		return ErrClosed
	}
	set := sp.subs[subject]
	subs := make([]*subscription, 0, len(set))
	for s := range set {
		subs = append(subs, s)
	}
	sp.mu.RUnlock()

	for _, s := range subs {
		s.enqueue(msg)
	}
	return nil
}

// Close отменяет подписки согласно ctx: ждет завершения горутин-обработчиков или возвращает по таймауту.
func (sp *subPub) Close(ctx context.Context) error {
	var err error
	sp.closeOnce.Do(func() {
		sp.mu.Lock()
		sp.closed = true

		all := make([]*subscription, 0)
		for _, set := range sp.subs {
			for s := range set {
				all = append(all, s)
			}
		}
		sp.subs = make(map[string]map[*subscription]struct{})
		sp.mu.Unlock()

		// сигналим всем подпискам на завершение loop
		for _, s := range all {
			s.unsubscribeInternal()
		}

		// ждем, пока все горутины обработчиков выйдут, или ctx истечет
		done := make(chan struct{})
		go func() {
			sp.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
			err = ctx.Err()
		}
	})
	return err
}

// enqueue добавляет сообщение в очередь подписчика и будит loop().
func (s *subscription) enqueue(msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.queue = append(s.queue, msg)
	s.cond.Signal()
}

// loop читает из очереди FIFO, пока подписка не закрыта.
func (s *subscription) loop() {
	for {
		s.mu.Lock()
		for len(s.queue) == 0 && !s.closed {
			s.cond.Wait()
		}
		if len(s.queue) == 0 && s.closed {
			s.mu.Unlock()
			return
		}
		msg := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()

		// вызываем callback без блокировок
		s.cb(msg)
	}
}

// Unsubscribe удаляет подписку из sp и сигналит loop() на выход.
func (s *subscription) Unsubscribe() {
	// удаляем из карты подписок
	s.sp.mu.Lock()
	if set, ok := s.sp.subs[s.subject]; ok {
		delete(set, s)
		if len(set) == 0 {
			delete(s.sp.subs, s.subject)
		}
	}
	s.sp.mu.Unlock()

	s.unsubscribeInternal()
}

// unsubscribeInternal сигналит о закрытии и разблокирует loop().
func (s *subscription) unsubscribeInternal() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.cond.Signal()
}
