package queuing

import (
	"context"
	"fmt"
	"testing"

	"github.com/xmonader/ewf"

	"github.com/redis/go-redis/v9"
)

// TestEnqueueDequeue tests enqueuing and dequeuing workflows from the queue
func TestEnqueueDequeue(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	store, err := ewf.NewSQLiteStore("test.db")
	if err != nil {
		t.Fatalf("store error: %v", err)
	}

	queue, err := NewRedisQueue(
		"testQueue",
		ewf.QueueOptions{AutoDelete: false},
		client,
	)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	t.Cleanup(func() {
		if err := queue.Close(context.Background()); err != nil {
			t.Errorf("failed to close queue: %v", err)
		}

		if err := store.Close(); err != nil {
			t.Errorf("failed to close store: %v", err)
		}
	})

	for i := 0; i < 5; i++ {
		workflow := ewf.NewWorkflow(fmt.Sprintf("test-workflow%d", i))
		err = queue.Enqueue(t.Context(), workflow)
		if err != nil {
			t.Fatalf("failed to enqueue workflow: %v", err)
		}
	}

	len, err := queue.Length(t.Context())
	if err != nil {
		t.Fatalf("failed to get queue length: %v", err)
	}
	// check queue length == 5
	if len != 5 {
		t.Fatalf("expected queue length 5, got %d", len)
	}

	// dequeue workflows and check they exist and their names (to check the dequeue order)
	for i := 0; i < 5; i++ {
		wf, err := queue.Dequeue(t.Context())
		if err != nil {
			t.Fatalf("failed to dequeue workflow: %v", err)
		}
		if wf == nil {
			t.Fatalf("expected workflow, got nil")
		}
		expectedName := fmt.Sprintf("test-workflow%d", i)
		if wf.Name != expectedName {
			t.Errorf("expected workflow name %s, got %s", expectedName, wf.Name)
		}
	}

	// check queue is empty now
	len, err = queue.Length(t.Context())
	if err != nil {
		t.Fatalf("failed to get queue length: %v", err)
	}
	if len != 0 {
		t.Errorf("expected queue length 0, got %d", len)
	}
}

// TestClose tests closing the queue and ensuring it is deleted from Redis
func TestClose(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	queue, err := NewRedisQueue(
		"testQueue",
		ewf.QueueOptions{AutoDelete: false},
		client,
	)
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	// enqueue an item for the queue to actually exist in redis
	err = queue.Enqueue(t.Context(), ewf.NewWorkflow("test-wf"))
	if err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	err = queue.Close(t.Context())
	if err != nil {
		t.Fatalf("failed to close queue: %v", err)
	}

	exists, err := queue.(*redisQueue).client.Exists(t.Context(), "testQueue").Result()
	if err != nil {
		t.Fatalf("failed to check queue existence: %v", err)
	}

	// check that the queue is deleted from redis
	if exists != 0 {
		t.Errorf("expected queue to be deleted, but it still exists")
	}

}

// TestQueueName tests passing different queue names and check for errors
func TestQueueName(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	t.Run("invalid queue name", func(t *testing.T) {
		q, err := NewRedisQueue(
			"test-queue!",
			ewf.QueueOptions{AutoDelete: false},
			client,
		)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if q != nil {
			t.Error("expected queue to be nil")
		}
	})

	t.Run("valid queue name with :/_", func(t *testing.T) {
		name := "deploy:users/123_queue:main"
		q, err := NewRedisQueue(
			name,
			ewf.QueueOptions{AutoDelete: false},
			client,
		)
		t.Cleanup(func() {
			if err := q.Close(context.Background()); err != nil {
				t.Errorf("failed to close queue: %v", err)
			}
		})
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}
		if q == nil {
			t.Error("expected queue to be created, got nil")
		}
		if q.Name() != name {
			t.Errorf("expected queue name to be %s, got %s", name, q.Name())
		}
	})

}