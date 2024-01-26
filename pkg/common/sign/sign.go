package sign

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jamesruan/sodium"
)

func VerifySign(msg, signature, publicKey string) bool {
	h := sha256.New()
	h.Write([]byte(msg))
	bs := h.Sum(nil)

	var m = sodium.Bytes(bs)

	pb, err := hex.DecodeString(publicKey)
	if err != nil {
		fmt.Println("err: ", err)
		return false
	}
	pubKey := sodium.SignPublicKey{Bytes: pb}

	sm, err := hex.DecodeString(signature)
	if err != nil {
		fmt.Println("err: ", err)
		return false
	}
	signedMessage := sodium.Signature{Bytes: sm}

	if err := m.SignVerifyDetached(signedMessage, pubKey); err != nil {
		return false
	}

	return true
}
