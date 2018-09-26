package simple

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Design Philosphy : Don't use channel for data passing. Instead use shared structure between go routines.

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

func (ht *HTTPTask) updateUmbrellaURI() (e error) {

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

	defer close(ht.NotifyC)
	defer close(ht.DoneC)
	defer close(ht.UpdateC)

	var e error

	if ht.Type == "UmbrellaReport" {
		ht.updateUmbrellaURI()
	}

	ht.Res, e = ht.Req.Send()
	ht.NotifyC <- e

	tc := time.NewTicker(ht.P.Interval).C
	hlc := time.NewTicker(ht.P.HowLong).C

	// TO DO: Timeout call before ticker kicks in
L:

	for {
		select {
		case timeNow := <-tc:
			fmt.Printf("Ticker kicked at %v %d \n ", timeNow, timeNow.Unix())
			if ht.Type == "UmbrellaReport" {
				ht.updateUmbrellaURI()
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
		case <-hlc:
			fmt.Printf("Task ran for duration %v seconds", ht.P.HowLong)
			ht.DoneC <- true

		}
	}

	//Don't execute again until first go routine has been executed.
	fmt.Println("Exit")

	return
}
