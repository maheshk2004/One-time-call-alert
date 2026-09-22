package workers

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type BackgroundScheduler struct {
	client *asynq.Client
	stopCh chan struct{}
}

func NewBackgroundScheduler(client *asynq.Client) *BackgroundScheduler {
	return &BackgroundScheduler{
		client: client,
		stopCh: make(chan struct{}),
	}
}

func (s *BackgroundScheduler) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		log.Printf("Starting Background Scheduler (interval: %v)", interval)
		// Run initial check
		s.enqueueOneCallCheck()

		for {
			select {
			case <-ticker.C:
				s.enqueueOneCallCheck()
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *BackgroundScheduler) Stop() {
	close(s.stopCh)
}

func (s *BackgroundScheduler) enqueueOneCallCheck() {
	task, err := NewDetectOneCallTask()
	if err != nil {
		log.Printf("Error creating detect one call task: %v", err)
		return
	}

	info, err := s.client.EnqueueContext(context.Background(), task, asynq.MaxRetry(3), asynq.Timeout(60*time.Second))
	if err != nil {
		log.Printf("Failed to enqueue detect one call task: %v", err)
		return
	}
	log.Printf("Enqueued DETECT_ONE_CALL task (ID: %s, Queue: %s)", info.ID, info.Queue)
}
