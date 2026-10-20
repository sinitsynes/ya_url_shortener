package service

import (
	"context"
	"sync"
	"uuid"
)

type toDelete struct {
	userID   *uuid.UUID
	shortURL string
}

func fanOut(resources []toDelete) []chan toDelete {
	chans := make([]chan toDelete, len(resources))
	for i, resource := range resources {
		ch := make(chan toDelete)
		chans[i] = ch
		go func(r toDelete) {
			ch <- r
			close(ch)
		}(resource)
	}
	return chans
}

func fanIn(ctx context.Context, deleteCh []chan toDelete) chan toDelete {
	resultCh := make(chan toDelete)
	var wg sync.WaitGroup

	for _, ch := range deleteCh {
		chClosure := ch
		wg.Go(func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-ctx.Done():
					return
				case resultCh <- data:
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	return resultCh
}
