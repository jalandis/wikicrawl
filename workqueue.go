package wikicrawl

import (
	"sync"
	"time"
)

type WorkQueue struct {
	crawler Crawler
	wait    sync.WaitGroup
	todo    chan Link
	Result  *CrawlResult
}

func (wq *WorkQueue) AddWork(href Link) {
	wq.wait.Add(1)
	for {
		select {
		case wq.todo <- href:
			return
		case <-time.After(5 * time.Second):
			panic("Queue full")
		}
	}
}

func (wq *WorkQueue) Start(pool int) {
	for i := 0; i < pool; i++ {
		go func() {
			for work := range wq.todo {
				func() {
					defer wq.wait.Done()
					wq.crawler.FollowLink(work, wq)
				}()
			}
		}()
	}
}

func (wq *WorkQueue) Wait() {
	wq.wait.Wait()
	close(wq.todo)
}

func NewWorkQueue(crawler Crawler, limit int) *WorkQueue {
	queue := new(WorkQueue)
	queue.crawler = crawler
	queue.todo = make(chan Link, limit)
	queue.Result = &CrawlResult{Visited: NewLinkSet(), Broken: NewLinkSet()}

	return queue
}
