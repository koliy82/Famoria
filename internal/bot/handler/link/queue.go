package link

import (
	"net/url"
	"sync"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

// job represents a single link-to-video processing task.
type job struct {
	update telego.Update
	url    *url.URL
}

// queueDepth bounds how many links one chat may have waiting. With a per-video
// deadline of ten minutes a deep queue would mean hours of backlog, so a chat
// that floods the bot with links has the excess dropped instead of accumulated.
const queueDepth = 16

// queueManager ensures that link processing for different chats runs in
// parallel, while jobs within a single chat are processed strictly
// sequentially (one at a time). This keeps one busy chat from overloading the
// bot while still allowing other chats to make progress.
type queueManager struct {
	mu     sync.Mutex
	queues map[int64]chan job
	worker func(j job)
	log    *zap.Logger
}

func newQueueManager(worker func(j job), log *zap.Logger) *queueManager {
	return &queueManager{
		queues: make(map[int64]chan job),
		worker: worker,
		log:    log,
	}
}

// submit enqueues a job for the given chat. The first submit for a chat lazily
// starts a dedicated worker goroutine that drains its channel sequentially.
//
// It never blocks the caller: when a chat's queue is full the job is dropped.
// Blocking here would stall the handler goroutine for as long as the backlog
// takes to drain, which with a ten-minute per-video deadline is unbounded.
func (q *queueManager) submit(chatID int64, j job) bool {
	q.mu.Lock()
	ch, ok := q.queues[chatID]
	if !ok {
		ch = make(chan job, queueDepth)
		q.queues[chatID] = ch
		go q.run(chatID, ch)
	}
	q.mu.Unlock()

	select {
	case ch <- j:
		return true
	default:
		q.log.Info("link: chat queue is full, dropping link",
			zap.Int64("chat_id", chatID), zap.Int("capacity", queueDepth))
		return false
	}
}

// run is the per-chat worker loop. It processes jobs one at a time in the order
// they were submitted.
//
// A panic in one job is recovered rather than allowed to kill this goroutine:
// without that the channel would stop draining and every later submit for this
// chat would be silently dropped forever.
func (q *queueManager) run(chatID int64, ch chan job) {
	for j := range ch {
		q.runOne(chatID, j)
	}
}

func (q *queueManager) runOne(chatID int64, j job) {
	defer func() {
		if rec := recover(); rec != nil {
			q.log.Error("link: panic while processing a link",
				zap.Int64("chat_id", chatID), zap.Any("panic", rec))
		}
	}()
	q.worker(j)
}
