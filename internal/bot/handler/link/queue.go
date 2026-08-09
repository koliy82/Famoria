package link

import (
	"net/url"
	"sync"

	"github.com/mymmrac/telego"
)

// job represents a single link-to-video processing task.
type job struct {
	update telego.Update
	url    *url.URL
}

// queueManager ensures that link processing for different chats runs in
// parallel, while jobs within a single chat are processed strictly
// sequentially (one at a time). This keeps one busy chat from overloading the
// bot while still allowing other chats to make progress.
type queueManager struct {
	mu     sync.Mutex
	queues map[int64]chan job
	worker func(j job)
}

func newQueueManager(worker func(j job)) *queueManager {
	return &queueManager{
		queues: make(map[int64]chan job),
		worker: worker,
	}
}

// submit enqueues a job for the given chat. The first submit for a chat lazily
// starts a dedicated worker goroutine that drains its channel sequentially.
func (q *queueManager) submit(chatID int64, j job) {
	q.mu.Lock()
	ch, ok := q.queues[chatID]
	if !ok {
		ch = make(chan job, 64)
		q.queues[chatID] = ch
		go q.run(chatID, ch)
	}
	q.mu.Unlock()
	ch <- j
}

// run is the per-chat worker loop. It processes jobs one at a time in the order
// they were submitted.
func (q *queueManager) run(chatID int64, ch chan job) {
	for j := range ch {
		q.worker(j)
	}
}
