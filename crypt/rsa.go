package crypt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

/*
 * RSA encrypt algorithm
 */

//face info
type Rsa struct {
	pubKey string
	prvKey string
}

//construct
func NewRsa() *Rsa {
	this := &Rsa{}
	return this
}

//demo code
func (f *Rsa) Demo() {
	//generate rsa keys
	fmt.Println("------------------------------Generate RSA Keys-----------------------------------------")
	prvKey, pubKey, _ := f.GenRsaKey()

	fmt.Println("-------------------------------Verify Signature-----------------------------------------")
	var data = `This is origin data`
	signData, _ := f.RsaSignWithSha256([]byte(data), prvKey)
	fmt.Println("data signature info： ", hex.EncodeToString(signData))
	err := f.RsaVerySignWithSha256([]byte(data), signData, pubKey)
	if err == nil {
		fmt.Println("signature check succeed")
	}

	fmt.Println("-------------------------------Decode Data-----------------------------------------")
	cipherText, _ := f.RsaEncrypt([]byte(data), pubKey)
	fmt.Println("Pub key enc data：", hex.EncodeToString(cipherText))
	sourceData, _ := f.RsaDecrypt([]byte(data), prvKey)
	fmt.Println("Prv key decode data：", string(sourceData))
}

//set key
func (f *Rsa) SetKey(pubKey, prvKey string) error {
	//check
	if pubKey == "" || prvKey == "" {
		return errors.New("invalid parameter")
	}
	f.pubKey = pubKey
	f.prvKey = prvKey
	return nil
}

//private key decode
func (f *Rsa) RsaDecrypt(cipherText, keyBytes []byte) ([]byte, error) {
	//get private key
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("private key error")
	}
	//parse PKCS1 format private key
	prvKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//decode
	data, err := rsa.DecryptPKCS1v15(rand.Reader, prvKey, cipherText)
	return data, err
}

//public key encode
func (f *Rsa) RsaEncrypt(data, keyBytes []byte) ([]byte, error) {
	//decode pem format public key
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("public key error")
	}
	//parse public key
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	//encrypt
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return nil, err
	}
	return cipherText, nil
}

//verify signature
func (f *Rsa) RsaVerySignWithSha256(data, signData, keyBytes []byte) error {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return errors.New("public key error")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	return err
}

//gen signature
func (f *Rsa) RsaSignWithSha256(data, keyBytes []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("private key error")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsePKCS8PrivateKey err:%v", err)
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return nil, fmt.Errorf("error from signing, err:%v", err)
	}
	return signature, nil
}

//gen rsa private and public key
//return prvKey, pubKey, error
func (f *Rsa) GenRsaKey() ([]byte, []byte, error) {
	var (
		prvKey, pubKey []byte
		err error
	)
	//gen private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey = pem.EncodeToMemory(block)
	//gen public key
	publicKey := &privateKey.PublicKey
	derPKix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPKix,
	}
	pubKey = pem.EncodeToMemory(block)
	return prvKey, pubKey, nil
}