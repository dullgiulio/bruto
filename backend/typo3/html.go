package typo3

import (
	"errors"
	"io"

	"golang.org/x/net/html"
)

var errRsaNotFound = errors.New("RSA Public Key not found")

func (e *encrypter) pkFromHTML(r io.Reader) error {
	z := html.NewTokenizer(r)
Loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			err := z.Err()
			if err == io.EOF {
				break Loop
			}
			return err
		case html.SelfClosingTagToken:
			tagName, hasAttr := z.TagName()
			tn := string(tagName)
			if !hasAttr || tn != "input" {
				continue
			}
			var isExp, isMod bool
			for {
				key, val, more := z.TagAttr()
				k := string(key)
				v := string(val)
				switch k {
				case "id":
					switch v {
					case "rsa_n":
						isMod = true
					case "rsa_e":
						isExp = true
					}
				case "value":
					if isExp {
						e.Exp = v
						continue
					}
					if isMod {
						e.Mod = v
						continue
					}
				}
				if !more {
					break
				}
			}
		}
	}
	if e.Mod == "" || e.Exp == "" {
		return errRsaNotFound
	}
	return nil
}
