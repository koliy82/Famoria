package waiter

import (
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

// WaitedMessage holds data about a single awaited message.
type WaitedMessage struct {
	OwnerID   int64
	ChatID    int64
	ReplyToID int // if >0, message must be a reply to this message ID
	Deadline  time.Time
	Callback  func(ctx *th.Context, msg telego.Message)
}

// MessageWaiter manages waiters using channels instead of mutexes.
// All state is owned by the internal goroutine that processes channels.
type MessageWaiter struct {
	addCh      chan WaitedMessage
	removeCh   chan int64
	incomingCh chan incomingMsg
	stopCh     chan struct{}
	log        *zap.Logger
}

type incomingMsg struct {
	ctx *th.Context
	msg telego.Message
}

func New(log *zap.Logger) *MessageWaiter {
	mw := &MessageWaiter{
		addCh:      make(chan WaitedMessage),
		removeCh:   make(chan int64),
		incomingCh: make(chan incomingMsg),
		stopCh:     make(chan struct{}),
		log:        log,
	}
	go mw.loop()
	return mw
}

func (mw *MessageWaiter) loop() {
	waiters := make(map[int64]WaitedMessage)
	for {
		select {
		case w := <-mw.addCh:
			// add or replace waiter
			waiters[w.OwnerID] = w
		case owner := <-mw.removeCh:
			delete(waiters, owner)
		case im := <-mw.incomingCh:
			m := im.msg
			w, ok := waiters[m.From.ID]
			if !ok {
				continue
			}
			// check chat
			if w.ChatID != 0 && w.ChatID != m.Chat.ID {
				continue
			}
			// check reply
			if w.ReplyToID > 0 {
				if m.ReplyToMessage == nil || m.ReplyToMessage.MessageID != w.ReplyToID {
					continue
				}
			}
			// remove waiter before calling to make it one-shot
			delete(waiters, m.From.ID)
			// call callback in goroutine
			go func(cb func(ctx *th.Context, msg telego.Message), ctx *th.Context, msg telego.Message) {
				defer func() {
					if r := recover(); r != nil {
						mw.log.Sugar().Errorf("panic in message waiter callback: %v", r)
					}
				}()
				cb(ctx, msg)
			}(w.Callback, im.ctx, m)
		case <-mw.stopCh:
			return
		}
	}
}

// WaitForMessage registers a one-shot waiter. Implementation uses channel to send add request.
func (mw *MessageWaiter) WaitForMessage(ownerID int64, chatID int64, replyToID int, duration time.Duration, cb func(ctx *th.Context, msg telego.Message)) {
	w := WaitedMessage{
		OwnerID:   ownerID,
		ChatID:    chatID,
		ReplyToID: replyToID,
		Deadline:  time.Now().Add(duration),
		Callback:  cb,
	}
	mw.addCh <- w
	// schedule removal after duration
	go func(uid int64, d time.Duration) {
		time.Sleep(d)
		select {
		case mw.removeCh <- uid:
		default:
		}
	}(ownerID, duration)
}

// HandleMessageUpdate should be used as message handler in bot handler.
func (mw *MessageWaiter) HandleMessageUpdate(ctx *th.Context, update telego.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}
	// send incoming message to manager goroutine
	mw.incomingCh <- incomingMsg{ctx: ctx, msg: *update.Message}
}

// Stop gracefully stops the internal loop.
func (mw *MessageWaiter) Stop() {
	close(mw.stopCh)
}
