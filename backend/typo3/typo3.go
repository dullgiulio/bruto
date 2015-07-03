package typo3

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/dullgiulio/bruto/backend"
	"github.com/dullgiulio/bruto/gen"
)

type T struct {
	// URLs configuration
	urls urls
	// session is ready.
	enc *encrypter
}

func New() *T {
	return &T{enc: &encrypter{}}
}

func (t *T) Setup(domain string, conn *backend.HTTP) {
	t.urls = urls(domain)
	// Preset constant POST values
	sp := conn.PostVals
	sp.Set("login_status", "login")
	sp.Set("redirect_url", "backend.php")
	sp.Set("loginRefresh", "")
	sp.Set("p_field", "")
	sp.Set("openid_url", "")
	sp.Set("commandLI", "Submit")
	sp.Set("interface", "backend")
}

func (t *T) Open(conn *backend.HTTP) error {
	// TODO: conn.Client.Timeout = ...
	req, err := conn.Get(t.urls.ajax())
	if err != nil {
		return err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	resp, err := conn.Do(req)
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
	if err := json.Unmarshal([]byte(xjson), &t.enc); err != nil {
		return err
	}
	return t.enc.seed()
}

func (t *T) Try(conn *backend.HTTP, l gen.Login) (success bool, err error) {
	// Encrypt password
	data, err := t.enc.encrypt(l.Pass)
	if err != nil {
		return
	}
	// Set request specific POST values
	conn.PostVals.Set("userident", data)
	conn.PostVals.Set("username", l.User)
	// Post login form
	req, err := conn.Post(t.urls.login(), &conn.PostVals)
	if err != nil {
		return
	}
	req.Header.Set("Referer", t.urls.referer())
	resp, err := conn.Do(req)
	if err != nil {
		return
	}
	// Don't care about body
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return false, errors.New(fmt.Sprintf("Invalid status code for login POST: %s", resp.Status))
	}
	// If the current location is "backend.php", we are in
	if strings.Index(resp.Request.URL.Path, "backend.php") >= 0 {
		success = true
	}
	return
}
