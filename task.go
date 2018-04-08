package simple

import (
	"fmt"
	"net/http"
	"time"
)

type Task interface {
	Run() (e error)
}

type HTTPTask struct {
	Req      HTTPReq
	Res      *http.Response
	Interval time.Duration
	HowLong  time.Duration
	DoneC    chan bool
	NotifyC  chan error
}

func (ht *HTTPTask) Run() (e error) {

	defer close(ht.NotifyC)
	defer close(ht.DoneC)

	ht.Res, e = ht.Req.Send()
	ht.NotifyC <- e

	tc := time.NewTicker(ht.Interval).C

	hlc := time.NewTicker(ht.HowLong).C

	// TO DO: Timeout call before ticker kicks in
L:

	for {
		select {
		case timeNow := <-tc:
			fmt.Println("Ticker kicked at ", timeNow)
			ht.Res, e = ht.Req.Send()
			ht.NotifyC <- e

		case dc := <-ht.DoneC:
			if dc {
				fmt.Println("Task done")
				break L
			}
			//Close response channel

		case <-hlc:
			fmt.Printf("Task ran for duration %v seconds", ht.HowLong)
			ht.DoneC <- true

		}
	}

	//Don't execute again until first go routine has been executed.
	fmt.Println("Exit")

	return
}
