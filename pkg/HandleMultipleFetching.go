package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RequestDetails struct {
	method    string
	link      string
	filepath  string
	TimeLimit time.Duration
}
type PossibleStatus string

const (
	Succesfull     PossibleStatus = "Succesfull"
	UknownnFailure PossibleStatus = "Failed for Unknown Reason"
	Timeout        PossibleStatus = "Ran out of Time"
	Cancelled      PossibleStatus = "User Cancelled"
)

type ResponseWithStatus struct {
	Response
	Error error
}

func RequestWorker(wg *sync.WaitGroup, inputChan <-chan RequestDetails, outputchan chan<- ResponseWithStatus, ctx context.Context) {
	localChan := make(chan RequestDetails, 1)
	localChanResponse := make(chan ResponseWithStatus, 1)
	go func() {
		for request := range localChan {
			response, err := EvaluateFetching(request.method, request.link, request.filepath, request.TimeLimit)
			ResponsesFilled := ResponseWithStatus{response, err}
			localChanResponse <- ResponsesFilled
		}
		defer close(localChanResponse)
	}()
	for request := range inputChan {
		localChan <- request
		select {
		case <-time.Tick(2 * time.Second): // will need to change this up later
			response := ResponseWithStatus{Response{}, fmt.Errorf("ran out of time!, took too long")}
			outputchan <- response
		case <-ctx.Done(): // means the users wants to stop the process!
			response := ResponseWithStatus{Response{}, fmt.Errorf("user Cancelled")}
			outputchan <- response
			return
			// pass it up as inconclusive
		case response := <-localChanResponse: // our request succesfully closed!
			outputchan <- response
		}
	}
	defer func() {
		wg.Done()
		close(localChan)
	}()
}

func RequestGenerator(wg *sync.WaitGroup, r RequestDetails, RequestPerSecond, CountTime int, inputChan chan<- RequestDetails, ctx context.Context) {
	defer wg.Done()
	done := make(chan bool)
	goodToProceedwithRequests := make(chan bool)
	go func() {
		HowManySecondHavePassed := 0
		defer func() {
			close(inputChan)
			done <- true
			close(done)
			close(goodToProceedwithRequests)
		}()

		for res := range goodToProceedwithRequests {
			if res {
				for range RequestPerSecond {
					inputChan <- r
				}
				HowManySecondHavePassed += 1
				if HowManySecondHavePassed == CountTime {
					return
				}
				time.Sleep(1 * time.Second)
			} else {
				return
			}
		}
	}()
	ticker := time.NewTicker(1 * time.Second)
	select {
	case <-ctx.Done():
		goodToProceedwithRequests <- false
	case <-ticker.C:
		goodToProceedwithRequests <- true
	case <-done:
		return
	}
}

func HandleOutput(outputChan chan ResponseWithStatus, CancelChan chan<- bool) {
	// eventually one would assume that
	Succesfull := 0
	Failed := 0
	for r := range outputChan {
		// I don't know for now print it !
		if r.IsOk {
			Succesfull += 1
		} else {
			Failed += 1
		}
		fmt.Printf("Succeded: %d out of %d\n", Succesfull, Succesfull+Failed)
	}
	// one that is done, let user know this is the same as user just cancelling
	CancelChan <- true
}

func RunAction(RequestPerSecond, CountTime int, r RequestDetails, workercount int, CancelChan chan bool) {
	var RequestGeneratorWaitGroup sync.WaitGroup
	var RequestWorkerWaitGroup sync.WaitGroup
	inputChan := make(chan RequestDetails, workercount)
	outputChan := make(chan ResponseWithStatus, 1)
	ctx, cancel := context.WithCancel(context.Background())

	RequestGeneratorWaitGroup.Add(1)
	go RequestGenerator(&RequestGeneratorWaitGroup, r, RequestPerSecond, CountTime, inputChan, ctx)
	for range workercount {
		RequestWorkerWaitGroup.Add(1)
		go RequestWorker(&RequestWorkerWaitGroup, inputChan, outputChan, ctx)
	}
	go func() {
		RequestWorkerWaitGroup.Wait()
		close(outputChan)
	}()
	go HandleOutput(outputChan, CancelChan)
	for <-CancelChan { // can be caused either by the end of HandleOutput OR by user sending one over!
		cancel()
	}

	defer func() {
		close(CancelChan)
	}()
}
