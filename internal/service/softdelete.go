package service

import (
	"context"
	"time"

	"ya_url_shortener/internal/model"
)

const (
	deleteBatchSize     = 50
	deleteFlushInterval = time.Second
)

func (c *Controller) deleteWorker() {
	buffer := make([]model.ToDelete, 0, deleteBatchSize)
	ticker := time.NewTicker(deleteFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		_ = c.store.SoftDeleteResources(context.Background(), buffer)
		buffer = buffer[:0]
	}

	for {
		select {
		case item, ok := <-c.deleteCh:
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, item)
			if len(buffer) == deleteBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
