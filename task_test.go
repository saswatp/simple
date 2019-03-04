package simple_test

import (
	"crypto/tls"
	"testing"
	"time"

	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/saswatp/simple"
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

	uri := "https://reports.api.umbrella.com/v1/organizations/2431158/security-activity?start=&stop="

	fmt.Println("Test start : Received URI ", uri)

	req := simple.HTTPReq{
		URI:           uri,
		Headers:       nil,
		ContentLength: 0,
		Method:        "GET",
		Body:          nil,
		RetryCount:    2,
		ShowDebug:     true,
		AP: simple.AuthParams{
			UserName: "4a31462421b4929bc6cf42f871c67ed",
			Password: "ccfd64025ed74615a363cc154068b5ab",
		},
		DialTimeout:         10 * time.Second,
		Timeout:             10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		KeepAlive:           0,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	p := simple.PollParams{
		Interval: 31 * time.Second,
		HowLong:  120 * time.Second,
	}

	task := simple.NewHTTPTask(req, p, simple.VariableUrlWithStartAndStop)

	//task.Run()

	var count int
	go task.Run()
L:
	for {

		select {

		case e, ok := <-task.NotifyC:

			if !ok {
				fmt.Println("Task notify channel closed")
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
				}
			}
			if count == 5 {
				task.DoneC <- true
			}
		}
	}
}

func TestHTTPTask_RunInvalidPoll(t *testing.T) {

	uri := "https://reports.api.umbrella.com/v1/organizations/2431158/security-activity?start=&stop="

	fmt.Println("Received URI ", uri)

	req := simple.HTTPReq{
		URI:           uri,
		Headers:       nil,
		ContentLength: 0,
		Method:        "GET",
		Body:          nil,
		RetryCount:    2,
		ShowDebug:     false,
		AP: simple.AuthParams{
			UserName: "34d43d3a6ec3463e9a5f5565716caa5d",
			Password: "da2a8e4204e84939a3c43449df44b93b",
		},
		DialTimeout:         10 * time.Second,
		Timeout:             20 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		KeepAlive:           0,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	p := simple.PollParams{
		Interval: 0 * time.Second,
		HowLong:  21 * time.Second,
	}

	task := simple.NewHTTPTask(req, p, "UmbrellaReport")
	//task.Run()

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
			}

		}
	}
}
