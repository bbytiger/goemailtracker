package emailtracker

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
	enc := &RSAEncoder{}
	enc.GeneratePrivateKey()
	return enc
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
