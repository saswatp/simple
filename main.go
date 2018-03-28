package simple

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	DefaultKeepAlive           time.Duration = 120 * time.Second
	DefaultDialTimeout         time.Duration = 0 * time.Second
	DefaultTimeout             time.Duration = 300 * time.Second
	DefaultTLSHandshakeTimeout time.Duration = 20 * time.Second

	ApplicationJSON string = "application/json"
	ContentType     string = "Content-Type"
)

type Requester interface {
	Send() (resp *http.Response, e error)
}

type HTTPReq struct {
	URI                 string
	Headers             map[string]string
	ContentLength       int64
	Method              string
	Body                []byte
	RetryCount          int
	ShowDebug           bool
	AP                  AuthParams
	DialTimeout         time.Duration
	Timeout             time.Duration
	TLSHandshakeTimeout time.Duration
	KeepAlive           time.Duration
	TLSConfig           *tls.Config

	// cookies           []*http.Cookie
	// QueryString       interface{}
	// Accept            string
	// Host              string
	// UserAgent         string
	// Insecure          bool
	// MaxRedirects      int
	// RedirectHeaders   bool
	// Proxy             string
	// Compression       *compression
	// CookieJar         http.CookieJar
	// OnBeforeRequest   func(goreq *Request, httpreq *http.Request)
}

type AuthParams struct {
	UserName string
	Password string
}

func IsJSON(s []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(s, &js) == nil
}

func (r *HTTPReq) Send() (resp *http.Response, e error) {

	var httpreq *http.Request
	httpreq, e = http.NewRequest(r.Method, r.URI, bytes.NewBuffer(r.Body))

	if e != nil {
		return
	}

	if len(r.Headers) != 0 {
		for key, value := range r.Headers {
			httpreq.Header.Set(key, value)
		}
	}
	//httpreq.Header.Set("content-type", "application/json")
	if r.ContentLength != 0 {
		httpreq.ContentLength = r.ContentLength
	}

	if r.AP.UserName != "" {
		httpreq.SetBasicAuth(r.AP.UserName, r.AP.Password)
	}

	dt := DefaultDialTimeout
	if r.DialTimeout != 0 {
		dt = r.DialTimeout
	}

	t := DefaultTimeout
	if r.Timeout != 0 {
		t = r.Timeout
	}

	ka := DefaultKeepAlive
	if r.KeepAlive != 0 {
		ka = r.KeepAlive
	}

	tht := DefaultTLSHandshakeTimeout
	if r.TLSHandshakeTimeout != 0 {
		tht = r.TLSHandshakeTimeout
	}

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   dt,
			KeepAlive: ka,
		}).Dial,
		TLSClientConfig:     r.TLSConfig,
		TLSHandshakeTimeout: tht,
		Proxy:               http.ProxyFromEnvironment,
	}

	//tr := &http.Transport{TLSClientConfig: r.TLSConfig}

	client := &http.Client{Transport: tr, Timeout: t}

	resp, e = client.Do(httpreq)
	return
}

// POST: Use only when you know that the return body is a string. Will not work for binary objects.
// When you're expecting a binary object use the Send function instead

func SendHTTPReq(req HTTPReq) (body []byte, e error) {

	resp, e := req.Send()
	if e != nil {
		e = fmt.Errorf("Error creating resource: %s", e.Error())
		return
	}
	body, e = ioutil.ReadAll(resp.Body)

	if e != nil {
		e = fmt.Errorf("Failed to read response from server %s", e.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		e = fmt.Errorf("Error: HTTP Status Code: %d \n Body: %s", resp.StatusCode, body)
		return
	}
	return

}

func POST(req HTTPReq) (body []byte, e error) {
	req.Method = "POST"
	if IsJSON(req.Body) {
		req.Headers[ContentType] = ApplicationJSON
	}

	body, e = SendHTTPReq(req)
	return
}

func PUT(req HTTPReq) (body []byte, e error) {
	req.Method = "PUT"
	if IsJSON(req.Body) {
		req.Headers[ContentType] = ApplicationJSON
	}

	body, e = SendHTTPReq(req)
	return
}

// GET: Use only when you know that the return body is a string. Will not work for binary objects.
// When you're downloading binary object use the Send function instead
func GET(req HTTPReq) (body []byte, e error) {

	//Download MUD File here and extract DACL information

	//mudResponse, e := common.DownloadFile(rep.MudURI, "/Users/saspraha/go-work/src/github.com/mud")
	req.Method = "GET"
	body, e = SendHTTPReq(req)
	return
}

func New() (h *HTTPReq) {
	h = &HTTPReq{}
	return
}
