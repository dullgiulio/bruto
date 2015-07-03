package bruto

import (
	"fmt"
	"time"
)

type urls string

const (
	urlReferer = "http://%s/typo3/"
	urlAjax    = "http://%s/typo3/ajax.php?ajaxID=BackendLogin%%3A%%3AgetRsaPublicKey&_dc=%d&skipSessionUpdate=1"
	urlLogin   = "http://%s/typo3/index.php"
)

func (u urls) ajax() string {
	return fmt.Sprintf(urlAjax, string(u), time.Now().Unix())
}

func (u urls) login() string {
	return fmt.Sprintf(urlLogin, string(u))
}

func (u urls) referer() string {
	return fmt.Sprintf(urlReferer, string(u))
}
