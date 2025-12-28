package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	TimeLimitForFetcher = 5 * time.Second
)

type RequestDetails struct {
	Method    string
	Link      string
	Filepath  string
	TimeLimit time.Duration
}
type PossibleStatus string

type TotalRequest struct {
	count int
	mu    sync.Mutex
}
type ResponseWithStatus struct {
	Response
	Error error
}

func (r *ResponseWithStatus) String() string {
	return fmt.Sprintf("ResponseWithStatus: Response: %v, Error: %v", r.Response, r.Error)
}

func RequestWorker(wg *sync.WaitGroup, inputChan <-chan RequestDetails, outputchan chan<- ResponseWithStatus, ctx context.Context) {
	defer func() {
		wg.Done()
	}()
	for request := range inputChan {
		localchan := make(chan ResponseWithStatus, 1)
		go func(localchan chan ResponseWithStatus, r RequestDetails) {
			res, err := EvaluateFetching(r.Filepath, r.Link, r.Filepath, TimeLimitForFetcher)
			localchan <- ResponseWithStatus{res, err}
		}(localchan, request)
		select {
		case <-ctx.Done():
			outputchan <- ResponseWithStatus{Response{
				request.Link, request.Method,
				make(map[string]any), nil,
				0 * time.Second, make(map[string]any), false,
			}, fmt.Errorf("user Cancelled")}
		case <-time.After(1 * time.Second):
			outputchan <- ResponseWithStatus{Response{
				request.Link, request.Method,
				make(map[string]any), nil,
				0 * time.Second, make(map[string]any), false,
			}, fmt.Errorf("took Too Much Time")}
		case response := <-localchan:
			outputchan <- response
		}
	}
}

func RequestGenerator(r RequestDetails, RequestPerSecond, CountTime int,
	inputChan chan<- RequestDetails, ctx context.Context, requestcount *TotalRequest,
) {
	defer func() {
		close(inputChan)
	}()
	Count := 0
	Ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-Ticker.C:
			for range RequestPerSecond {
				select {
				case <-ctx.Done():
					return
				case inputChan <- r:
					requestcount.mu.Lock()
					requestcount.count++
					requestcount.mu.Unlock()
				}
			}
			Count += 1
			if Count == CountTime {
				return
			}
		}
	}
}

func HandleOutput(wg *sync.WaitGroup, outputChan chan ResponseWithStatus, CancelChan chan<- bool) {
	// eventually one would assume that
	defer wg.Done()
	Succesfull := 0
	Failed := 0
	for r := range outputChan {
		// I don't know for now print it !
		if r.IsOk {
			Succesfull += 1
		} else if !r.IsOk || r.Error != nil {
			Failed += 1
		}
	}
	// one that is done, let user know this is the same as user just cancelling
	CancelChan <- true
}

func RunAction(RequestPerSecond, CountTime int, r RequestDetails, workercount int, CancelChan chan bool) {
	totalCount := TotalRequest{0, sync.Mutex{}}
	var RequestWorkerWaitGroup sync.WaitGroup
	var ProcessOutputWaitGroup sync.WaitGroup
	defer func() {
		close(CancelChan)
		fmt.Printf("total count is %v\n", totalCount.count)
	}()
	inputChan := make(chan RequestDetails, workercount)
	outputChan := make(chan ResponseWithStatus, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go RequestGenerator(r, RequestPerSecond, CountTime, inputChan, ctx, &totalCount)
	for range workercount {
		RequestWorkerWaitGroup.Add(1)
		go RequestWorker(&RequestWorkerWaitGroup, inputChan, outputChan, ctx)
	}
	go func() {
		RequestWorkerWaitGroup.Wait()
		close(outputChan)
	}()
	ProcessFinishedChan := make(chan bool, 1)
	ProcessOutputWaitGroup.Add(1)
	go HandleOutput(&ProcessOutputWaitGroup, outputChan, ProcessFinishedChan)
	select {
	case <-CancelChan: // user choose to end it
		fmt.Println("user choose to end it")
		cancel()
		ProcessOutputWaitGroup.Wait()
	case <-ProcessFinishedChan: // process ended by itself
		fmt.Println("we ended it")
		cancel()
	}
}
