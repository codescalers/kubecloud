package utils

import (
	"time"

	"github.com/xmonader/ewf"
)

func GetChaintQueueConfig() *ewf.QueueMetadata {
	return &ewf.QueueMetadata{
		Name: "chain_operations_queue",
		WorkersDef: ewf.WorkersDefinition{
			Count:        1,
			PollInterval: 1 * time.Second,
			WorkTimeout:  10 * time.Minute,
		},
		QueueOptions: ewf.QueueOptions{
			AutoDelete:  true,
			DeleteAfter: 10 * time.Minute,
			PopTimeout:  1 * time.Second,
		},
	}
}
