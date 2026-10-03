package link

import (
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

func testLogger(t *testing.T) *zap.Logger {
	t.Helper()
	log, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	return log
}

func dummyJob() job {
	return job{
		update: telego.Update{Message: &telego.Message{Chat: telego.Chat{ID: 1}}},
		url:    &url.URL{Scheme: "https", Host: "youtu.be", Path: "/x"},
	}
}

// TestSubmitDoesNotBlockWhenFull is the regression guard for the unbounded
// stall: submit used to block on a full channel, which would freeze the handler
// goroutine for as long as the backlog took to drain — hours, given the
// per-video deadline.
func TestSubmitDoesNotBlockWhenFull(t *testing.T) {
	// The worker blocks until released, so the queue fills up behind it.
	release := make(chan struct{})
	started := make(chan struct{}, 1)

	var once sync.Once
	q := newQueueManager(func(job) {
		once.Do(func() { close(started) })
		<-release
	}, testLogger(t))

	// First job occupies the worker, then the rest fill the buffered channel.
	// Waiting for the worker to actually pick up job one matters: without it the
	// buffer still holds that job and the count below is off by one.
	if !q.submit(7, dummyJob()) {
		t.Fatal("first submit should have been accepted")
	}
	<-started

	for i := 1; i < queueDepth+1; i++ {
		if !q.submit(7, dummyJob()) {
			t.Fatalf("submit %d should have been accepted (buffer holds %d)", i, queueDepth)
		}
	}

	// Beyond capacity, submit must return false promptly instead of blocking.
	done := make(chan bool, 1)
	go func() { done <- q.submit(7, dummyJob()) }()

	select {
	case accepted := <-done:
		if accepted {
			t.Error("submit on a full queue reported success")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("submit blocked on a full queue instead of dropping the job")
	}

	close(release)
}

// TestPanickingWorkerKeepsQueueAlive guards the second reliability bug: a panic
// in one job used to kill the worker goroutine, after which that chat's channel
// was never drained again and every later submit was dropped forever.
func TestPanickingWorkerKeepsQueueAlive(t *testing.T) {
	var mu sync.Mutex
	var processed []int
	release := make(chan int, queueDepth+4)

	q := newQueueManager(func(j job) {
		n := int(j.update.Message.Chat.ID)
		mu.Lock()
		processed = append(processed, n)
		mu.Unlock()
		if n == 1 {
			panic("simulated failure")
		}
		release <- n
	}, testLogger(t))

	const chatID = int64(42)

	q.submit(chatID, jobWithID(1)) // panics
	q.submit(chatID, jobWithID(2)) // must still be processed
	q.submit(chatID, jobWithID(3)) // and this one too

	deadline := time.After(5 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case <-release:
		case <-deadline:
			mu.Lock()
			t.Fatalf("worker died after the panic; processed only %v", processed)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(processed) != 3 {
		t.Errorf("processed %v, want all three jobs including the panicking one", processed)
	}
}

func jobWithID(chatID int64) job {
	return job{
		update: telego.Update{Message: &telego.Message{Chat: telego.Chat{ID: chatID}}},
		url:    &url.URL{Scheme: "https", Host: "example.com"},
	}
}

// TestQueuesArePerChat verifies the core design: one chat's backlog must not
// block another chat's jobs, since each chat gets its own worker.
func TestQueuesArePerChat(t *testing.T) {
	block := make(chan struct{})
	chatAStarted := make(chan struct{}, 1)
	chatBDone := make(chan struct{}, 1)

	q := newQueueManager(func(j job) {
		switch j.update.Message.Chat.ID {
		case 100:
			chatAStarted <- struct{}{}
			<-block // chat A is slow
		case 200:
			chatBDone <- struct{}{}
		}
	}, testLogger(t))

	q.submit(100, jobWithID(100))
	<-chatAStarted

	// Chat B must proceed while chat A is stuck on a long download.
	q.submit(200, jobWithID(200))
	select {
	case <-chatBDone:
	case <-time.After(3 * time.Second):
		t.Error("chat B was blocked by chat A's job; queues are not isolated per chat")
	}

	close(block)
}

// TestJobsWithinChatAreSequential asserts the other half of the design: a single
// chat must not have two videos downloading at once.
func TestJobsWithinChatAreSequential(t *testing.T) {
	var mu sync.Mutex
	inFlight := 0
	maxInFlight := 0
	done := make(chan struct{}, 3)

	q := newQueueManager(func(job) {
		mu.Lock()
		inFlight++
		if inFlight > maxInFlight {
			maxInFlight = inFlight
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		inFlight--
		mu.Unlock()
		done <- struct{}{}
	}, testLogger(t))

	for i := 0; i < 3; i++ {
		q.submit(555, dummyJob())
	}
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("jobs did not all complete")
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if maxInFlight != 1 {
		t.Errorf("max concurrent jobs in one chat = %d, want 1", maxInFlight)
	}
}
