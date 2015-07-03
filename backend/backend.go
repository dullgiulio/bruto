package backend

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
)

// TODO: This comes from a generator
const userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:38.0) Gecko/20100101 Firefox/38.0"

type HTTP struct {
	// Session connection
	Client *http.Client
	// Convenience shared POST values
	PostVals url.Values
	// Where to write debug dumps, nil for no debug
	debug io.Writer
}

func NewHTTP() *HTTP {
	return &HTTP{
		PostVals: url.Values{},
		Client:   &http.Client{},
	}
}

func (h *HTTP) Init() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	h.Client.Jar = jar
	return nil
}

func (h *HTTP) prepareReq(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	for _, cookie := range h.Client.Jar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}
	return req
}

func (h *HTTP) Do(req *http.Request) (*http.Response, error) {
	req = h.prepareReq(req)
	if h.debug != nil {
		b, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(h.debug, "%s\n", b)
	}
	res, err := h.Client.Do(req)
	if h.debug != nil {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(h.debug, "%s\n", b)
	}
	return res, err
}

func (h *HTTP) Post(url string, vals *url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func (h *HTTP) Get(url string) (*http.Request, error) {
	return http.NewRequest("GET", url, nil)
}
