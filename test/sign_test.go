package test

import (
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/sign"
	"testing"
)

func Test_SignVerify(t *testing.T) {
	msg := "abc"
	publicKey := "a4465fd76c16fcc458448076372abf1912cc5b150663a64dffefe550f96feadd"
	signature := "c09be8ad8b894cb05cf2e1ad3573680f85995636c083f9b564caa29905759d19735d63d503f3745ad3fa649ea2040fff32436f4bf5bfb172b4fcbff45b22f50b"
	if !sign.VerifySign(msg, signature, publicKey) {
		t.Fatal("signature forged")
	}

	msg = "forged msg"
	if sign.VerifySign(msg, signature, publicKey) {
		t.Fatal("signature forged")
	}
}
