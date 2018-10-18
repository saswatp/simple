package simple

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Design Philosphy : Don't use channel for data passing. Instead use shared structure between go routines.

const (
	fiveYears                   = 157680000
	VariableUrlWithStartAndStop = "variableUrlWithStartAndStop"
)

type Task interface {
	Run() (e error)
}

type HTTPTask struct {
	Type    string
	Req     HTTPReq
	Res     *http.Response
	P       PollParams
	DoneC   chan bool
	NotifyC chan error
	UpdateC chan HTTPReq
}

type PollParams struct {
	Interval time.Duration
	HowLong  time.Duration
	Count    int
}

func NewHTTPTask(req HTTPReq, p PollParams, dataType string) (ht *HTTPTask) {
	ht = &HTTPTask{
		Type: dataType,
		Req:  req,
		P:    p,
	}
	ht.DoneC = make(chan bool)
	ht.NotifyC = make(chan error)
	ht.UpdateC = make(chan HTTPReq)
	return
}

func (ht *HTTPTask) updateURI() (e error) {

	var newURL *url.URL
	var endTime, startTime int64

	//fmt.Printf("New UrI = %s", ht.Req.URI)
	if newURL, e = url.Parse(ht.Req.URI); e != nil {
		return
	}

	endTimeStr := newURL.Query().Get("end")
	if endTimeStr == "" {
		endTime = time.Now().Unix()
		startTime = endTime - int64(ht.P.Interval.Seconds())
	} else {
		endTime, _ = strconv.ParseInt(endTimeStr, 10, 64)
		startTime = endTime
		endTime = endTime + int64(ht.P.Interval.Seconds())
	}

	q := newURL.Query()

	q.Set("start", strconv.FormatInt(startTime, 10))
	q.Set("end", strconv.FormatInt(endTime, 10))

	newURL.RawQuery = q.Encode()
	ht.Req.URI = newURL.String()

	return
}

func (ht *HTTPTask) Run() {

	var e error
	var hlc *time.Ticker

	defer close(ht.NotifyC)
	defer close(ht.DoneC)
	defer close(ht.UpdateC)

	// To Do: Check if http timeout is greater than poll interval. If yes, fail the request
	if ht.P.HowLong <= 0 {
		// Run indefinitely - for 5 years
		ht.P.HowLong = time.Duration(fiveYears) * time.Second
	}
	if ht.Req.DialTimeout > ht.P.HowLong || ht.Req.Timeout > ht.P.HowLong || ht.Req.TLSHandshakeTimeout > ht.P.HowLong {
		e = errors.New("Invalid parameter - All timeout values must be less than poll interval ")
		ht.NotifyC <- e
		return
	}
	fmt.Printf("If not inturptted , this task will run for duration %v at interval %v \n", ht.P.HowLong, ht.P.Interval)

	fmt.Printf("Task started at time %v \n", time.Now())
	hlc = time.NewTicker(ht.P.HowLong)
	defer hlc.Stop()

	if ht.Type == VariableUrlWithStartAndStop {
		ht.updateURI()
	}
	fmt.Println("Sending first request to ", ht.Req.URI)
	ht.Res, e = ht.Req.Send()
	ht.NotifyC <- e

	if ht.P.Interval <= 0 {
		return
	}
	tc := time.NewTicker(ht.P.Interval)
	defer tc.Stop()

	// TO DO: Timeout call before ticker kicks in
L:

	for {
		select {
		case timeNow := <-tc.C:
			fmt.Printf("Ticker kicked at %v %d \n ", timeNow, timeNow.Unix())
			if ht.Type == VariableUrlWithStartAndStop {
				ht.updateURI()
			}
			fmt.Printf("Requesting URL %s\n", ht.Req.URI)
			ht.Res, e = ht.Req.Send()
			ht.NotifyC <- e

		case dc := <-ht.DoneC:
			if dc {
				fmt.Println("Task done")
				break L
			}
			//Close response channel

		case uc := <-ht.UpdateC:
			ht.Req = uc
		case <-hlc.C: //How long count

			break L
		}
	}
	fmt.Printf("Task ended at time %v \n", time.Now())
	//Don't execute again until first go routine has been executed.
	fmt.Println("Exit")

	return
}
