package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	//以太坊加密库，要求go版本升级到1.15
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func genPrivateKey() (*ecies.PrivateKey, error) {
	pubkeyCurve := elliptic.P256() //初始化椭圆曲线
	//随机挑选基点，生成私钥
	p, err := ecdsa.GenerateKey(pubkeyCurve, rand.Reader) //用golang标准库生成公私钥
	if err != nil {
		return nil, err
	} else {
		return ecies.ImportECDSA(p), nil //转换成以太坊的公私钥对
	}
}

// ECCEncrypt 椭圆曲线加密
func ECCEncrypt(plain string, pubKey *ecies.PublicKey) ([]byte, error) {
	src := []byte(plain)
	return ecies.Encrypt(rand.Reader, pubKey, src, nil, nil)
}

// ECCDecrypt 椭圆曲线解密
func ECCDecrypt(cipher []byte, prvKey *ecies.PrivateKey) (string, error) {
	if src, err := prvKey.Decrypt(cipher, nil, nil); err != nil {
		return "", err
	} else {
		return string(src), nil
	}
}

func main() {
	prvKey, err := genPrivateKey()
	if err != nil {
		fmt.Println(err)
	}
	pubKey := prvKey.PublicKey
	plain := "我们没什么不同"
	cipher, err := ECCEncrypt(plain, &pubKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("密文：%v\n", cipher)
	plain, err = ECCDecrypt(cipher, prvKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("明文：%s\n", plain)
}

//>>>>>>>>>>output
//密文：[4 227 99 4 70 47 3 171 28 35 201 222 67 129 156 104 187 76 234 53 88 162 183 89 77 165 16 247 79 172 25 104 26 116 120 6 62 38 71 252 224 137 207 224 143 188 157 47 202 18 127 57 104 24 155 75 222 166 133 20 235 77 231 78 174 210 52 31 69 100 3 20 131 148 179 101 172 0 163 231 242 188 189 198 206 9 125 137 247 44 126 38 249 150 63 49 61 59 132 40 209 28 84 18 5 76 78 141 233 196 111 186 247 4 48 172 50 227 109 63 232 225 151 249 3 36 126 197 236 100 175 214 170 64]
//明文：我们没什么不同
