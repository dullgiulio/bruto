package typo3

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strconv"
)

type encrypter struct {
	rand io.Reader
	pk   rsa.PublicKey
	// For JSON unmarshalling
	Mod string `json:"publicKeyModulus"`
	// For JSON unmarshalling
	Exp string `json:"exponent"`
}

func (e *encrypter) seed() (err error) {
	var mod big.Int
	var exp int64
	if _, ok := mod.SetString(e.Mod, 16); !ok {
		return errors.New(fmt.Sprintf("Cannot parse MOD from hexadecimal value: %s", e.Mod))
	}
	if exp, err = strconv.ParseInt(e.Exp, 16, 0); err != nil {
		return
	}
	e.pk.N = &mod
	e.pk.E = int(exp)
	return
}

func (e *encrypter) encrypt(pass string) (b string, err error) {
	if e.rand == nil {
		e.rand = rand.Reader
	}
	var data []byte
	if data, err = rsa.EncryptPKCS1v15(e.rand, &e.pk, []byte(pass)); err != nil {
		return
	}
	return "rsa:" + base64.StdEncoding.EncodeToString(data), nil
}
