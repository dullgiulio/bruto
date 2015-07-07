package backend

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
)

type HTTP struct {
	// User-provided headers
	Header http.Header
	// Session connection
	Client *http.Client
	// Convenience shared POST values
	PostVals url.Values
	// Where to write debug dumps, nil for no debug
	debug io.Writer
}

func NewHTTP() *HTTP {
	return &HTTP{
		Header:   http.Header(make(map[string][]string)),
		PostVals: url.Values{},
		Client:   &http.Client{},
		//	debug:    os.Stdout,
	}
}

func (h *HTTP) Init() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	h.Client.Jar = jar
	h.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		h.prepareReq(req)
		return nil
	}
	return nil
}

func (h *HTTP) prepareReq(req *http.Request) *http.Request {
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	for _, cookie := range h.Client.Jar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}
	// Override defaults with request specific headers
	for k, v := range h.Header {
		// XXX: Using only one value per key
		req.Header.Set(k, v[0])
	}
	return req
}

func (h *HTTP) Do(req *http.Request) (*http.Response, error) {
	h.prepareReq(req)
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
