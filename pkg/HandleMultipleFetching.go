package pkg

import (
	"context"
	"sync"
	"time"
)

type RequestDetails struct {
	method    string
	link      string
	filepath  string
	TimeLimit time.Duration
}

func RequestWorker(wg *sync.WaitGroup, inputChan <-chan RequestDetails, outputchan chan<- RequestDetails, ctx context.Context) {
	defer wg.Done()
	localChan := make(chan RequestDetails, 1)
	go func() {
		for request := range localChan {
			response, err := EvaluateFetching(request.method, request.link, request.filepath, request.TimeLimit)
		}
	}()
	for request := range inputChan {
		localChan <- request
		select {
		case <-time.Tick(2 * time.Second): // will need to change this up later
		// failure just pass it up AS Failed
		case <-ctx.Done():
			// pass it up as inconclusive
		}
	}
}

func RunAction(RequestPerSecond int, method, link, bodyfile string, workercount int) {
	var RequestWorkerWaitGroup sync.WaitGroup
	for eachrequest := range RequestPerSecond {
		RequestWorkerWaitGroup.Add(1)
	}
}
