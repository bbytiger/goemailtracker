package emailtracker

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
)

// extend to implement your own encoding/encryption
type Encoder interface {
	EncodeMessage(strMsg string) []byte
	DecodeMessage(encodedByteMsg []byte) string
}

// example rsa encryption implementation
type RSAEncoder struct {
	privKey *rsa.PrivateKey
}

func NewRSAEncoder() *RSAEncoder {
	return &RSAEncoder{}
}

func (d *RSAEncoder) LoadPublicPrivateKeys(pub []byte, priv []byte) {
	pubBlock, _ := pem.Decode(pub)
	privBlock, _ := pem.Decode(priv)
	pubKey, pubErr := x509.ParsePKCS1PublicKey(pubBlock.Bytes)
	privKey, privErr := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if pubErr != nil {
		panic(pubErr)
	}
	if privErr != nil {
		panic(pubErr)
	}
	privKey.PublicKey = *pubKey
}

func (d *RSAEncoder) GeneratePrivateKey() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	d.privKey = privateKey
}

func (d RSAEncoder) EncodeMessage(strMsg string) []byte {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(), rand.Reader, &d.privKey.PublicKey, []byte(strMsg), nil,
	)
	if err != nil {
		panic(err)
	}
	return encryptedBytes
}

func (d RSAEncoder) DecodeMessage(encodedByteMsg []byte) string {
	decryptedBytes, err := d.privKey.Decrypt(
		nil, encodedByteMsg, &rsa.OAEPOptions{Hash: crypto.SHA256},
	)
	if err != nil {
		panic(err)
	}
	return string(decryptedBytes)
}
