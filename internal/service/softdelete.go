package service

import (
	"context"

	"ya_url_shortener/internal/model"
)

const deleteBatchSize int = 50

func (c *Controller) deleteWorker() {
	buffer := make([]model.ToDelete, 0, deleteBatchSize)
	flush := func() {
		if len(buffer) == 0 {
			return
		}
		_ = c.store.SoftDeleteResources(context.Background(), buffer)
		buffer = buffer[:0]
	}
	for item := range c.deleteCh {
		buffer = append(buffer, item)
		if len(buffer) == deleteBatchSize {
			flush()
		}
	}
	flush()
}
