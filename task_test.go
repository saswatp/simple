package simple_test

import (
	"crypto/tls"
	"testing"
	"time"

	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/simple"
)

func TestSend(t *testing.T) {
	t.Skip()

	req := simple.HTTPReq{
		URI:       "https://google.com",
		Method:    "GET",
		TLSConfig: &tls.Config{},
	}

	resp, e := req.Send()
	defer resp.Body.Close()

	if e != nil {
		t.Log(e.Error())

	} else {

		if resp.StatusCode < http.StatusBadRequest {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			t.Log(string(bodyBytes))
		}

	}
	return
}

func TestHTTPTask_Run(t *testing.T) {

	notifyC := make(chan error)
	doneC := make(chan bool)
	//Create a task

	/*task := simple.HTTPTask{
		Req: simple.HTTPReq{
			URI:                 "
			Headers:             nil,
			ContentLength:       0,
			Method:              "GET",
			Body:                nil,
			RetryCount:          0,
			ShowDebug:           false,
			AP:                  simple.AuthParams{},
			DialTimeout:         2 * time.Second,
			Timeout:             0,
			TLSHandshakeTimeout: 0,
			KeepAlive:           0,
			TLSConfig:           &tls.Config{},
		},
		Interval: 5 * time.Second,
		HowLong:  30 * time.Second,
		DoneC:    doneC,
		NotifyC:  notifyC,
	}
	*/

	task := simple.HTTPTask{
		Req: simple.HTTPReq{
			URI:           "https://reports.api.umbrella.com/v1/organizations/2431158/security-activity",
			Headers:       nil,
			ContentLength: 0,
			Method:        "GET",
			Body:          nil,
			RetryCount:    2,
			ShowDebug:     false,
			AP: simple.AuthParams{
				UserName: "ac07755a9de146e78f452f65d9494fd2",
				Password: "072c37da97ef40e5a0d6faa89f362f22",
			},
			DialTimeout:         10 * time.Second,
			Timeout:             60 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			KeepAlive:           0,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Interval: 5 * time.Second,
		HowLong:  30 * time.Second,
		DoneC:    doneC,
		NotifyC:  notifyC,
	}
	var count int
	go task.Run()

L:
	for {

		select {

		case e, ok := <-task.NotifyC:

			if !ok {
				break L
			}
			count++
			if e != nil {
				fmt.Println("Error = ", e.Error())
			} else {
				if task.Res.StatusCode < http.StatusBadRequest {
					bodyBytes, _ := ioutil.ReadAll(task.Res.Body)
					fmt.Println("Response = ", string(bodyBytes))
					task.Res.Body.Close()

				} else {
					t.Fail()
				}
			}

			if count == 5 {
				task.DoneC <- true
				break L
			}

		case _, ok := <-task.DoneC:
			if !ok {
				break L
			}

		}

	}

}
