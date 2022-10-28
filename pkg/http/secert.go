package http

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"

	"github.com/teyvat-helper/hk4e-emu/pkg/ec2b"
)

type Secret struct {
	Shared *ec2b.Ec2b
	Server *PrivateKey
	Client map[string]*PublicKey
}

func NewSecret() *Secret {
	s := &Secret{}
	s.Server = &PrivateKey{}
	s.Client = make(map[string]*PublicKey)
	rest, _ := os.ReadFile("data/secret.pem")
	var block *pem.Block
	for {
		block, rest = pem.Decode(rest)
		switch block.Type {
		case "DISPATCH SERVER RSA PRIVATE KEY":
			s.Server.PrivateKey, _ = x509.ParsePKCS1PrivateKey(block.Bytes)
		case "DISPATCH CLIENT RSA PUBLIC KEY 2":
			k, _ := x509.ParsePKCS1PublicKey(block.Bytes)
			s.Client["2"] = &PublicKey{k}
		case "DISPATCH CLIENT RSA PUBLIC KEY 3":
			k, _ := x509.ParsePKCS1PublicKey(block.Bytes)
			s.Client["3"] = &PublicKey{k}
		}
		if len(rest) == 0 {
			break
		}
	}
	return s
}

type PrivateKey struct {
	*rsa.PrivateKey
}

func (k *PrivateKey) Sign(msg []byte) ([]byte, error) {
	hasher := sha256.New()
	hasher.Write(msg)
	digest := hasher.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, k.PrivateKey, crypto.SHA256, digest)
}

func (k *PrivateKey) SignBase64(msg []byte) (string, error) {
	sign, err := k.Sign(msg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sign), nil
}

func (k *PrivateKey) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, k.PrivateKey, ciphertext)
}

func (k *PrivateKey) DecryptBase64(s string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return k.Decrypt(ciphertext)
}

type PublicKey struct {
	*rsa.PublicKey
}

func (k *PublicKey) Encrypt(msg []byte) ([]byte, error) {
	var block, out []byte
	var err error
	size := k.Size() - 11
	for len(msg) > 0 {
		if len(msg) > size {
			block = msg[:size]
			msg = msg[size:]
		} else {
			block = msg
			msg = nil
		}
		block, err = rsa.EncryptPKCS1v15(rand.Reader, k.PublicKey, block)
		if err != nil {
			return nil, err
		}
		out = append(out, block...)
	}
	return out, nil
}

func (k *PublicKey) EncryptBase64(msg []byte) (string, error) {
	ciphertext, err := k.Encrypt(msg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}