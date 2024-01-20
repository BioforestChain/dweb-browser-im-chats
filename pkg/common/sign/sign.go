package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"strings"
)

func GenKeyPair() (privateKey string, publicKey string, e error) {
	priKey, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		return "", "", e
	}
	ecPrivateKey, e := x509.MarshalECPrivateKey(priKey)
	if e != nil {
		return "", "", e
	}
	privateKey = base64.StdEncoding.EncodeToString(ecPrivateKey)

	X := priKey.X
	Y := priKey.Y
	xStr, e := X.MarshalText()
	if e != nil {
		return "", "", e
	}
	yStr, e := Y.MarshalText()
	if e != nil {
		return "", "", e
	}
	public := string(xStr) + "+" + string(yStr)
	publicKey = base64.StdEncoding.EncodeToString([]byte(public))
	return
}

func BuildPrivateKey(privateKeyStr string) (priKey *ecdsa.PrivateKey, e error) {
	bytes, e := base64.StdEncoding.DecodeString(privateKeyStr)
	if e != nil {
		return nil, e
	}
	priKey, e = x509.ParseECPrivateKey(bytes)
	if e != nil {
		return nil, e
	}
	return
}

func BuildPublicKey(publicKeyStr string) (pubKey *ecdsa.PublicKey, e error) {
	bytes, e := base64.StdEncoding.DecodeString(publicKeyStr)
	if e != nil {
		return nil, e
	}
	split := strings.Split(string(bytes), "+")
	xStr := split[0]
	yStr := split[1]
	x := new(big.Int)
	y := new(big.Int)
	e = x.UnmarshalText([]byte(xStr))
	if e != nil {
		return nil, e
	}
	e = y.UnmarshalText([]byte(yStr))
	if e != nil {
		return nil, e
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	pubKey = &pub
	return
}

func Sign(content string, privateKeyStr string) (signature string, e error) {
	priKey, e := BuildPrivateKey(privateKeyStr)
	if e != nil {
		return "", e
	}
	r, s, e := ecdsa.Sign(rand.Reader, priKey, []byte(hash(content)))
	if e != nil {
		return "", e
	}
	rt, e := r.MarshalText()
	st, e := s.MarshalText()
	signStr := string(rt) + "+" + string(st)
	signature = hex.EncodeToString([]byte(signStr))
	return
}

func VerifySign(content, signature, publicKeyStr string) bool {
	decodeSign, e := hex.DecodeString(signature)
	if e != nil {
		return false
	}
	split := strings.Split(string(decodeSign), "+")
	rStr := split[0]
	sStr := split[1]
	rr := new(big.Int)
	ss := new(big.Int)
	e = rr.UnmarshalText([]byte(rStr))
	e = ss.UnmarshalText([]byte(sStr))
	pubKey, e := BuildPublicKey(publicKeyStr)
	if e != nil {
		return false
	}
	return ecdsa.Verify(pubKey, []byte(hash(content)), rr, ss)
}

// Hash算法，这里是sha256，可以根据需要自定义
func hash(data string) string {
	sum := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(sum[:])
}
