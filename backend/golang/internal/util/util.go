package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	mrand "math/rand"
	"slices"
)

func CreateRandStr(length int) (string, error) {
	// 指定した長さのバイトを生成
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// URLでも使えるランダムな文字列に
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Encrypt(data string, key []byte) (string, error) {
	cipherText := make([]byte, aes.BlockSize+len([]byte(data)))
	nonce := cipherText[:aes.BlockSize]
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	encryptStream := cipher.NewCTR(block, nonce)
	encryptStream.XORKeyStream(cipherText[aes.BlockSize:], []byte(data))
	return base64.RawURLEncoding.EncodeToString(cipherText), nil
}

func Decrypt(text string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	byteText, err := base64.RawURLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}
	nonce := byteText[:aes.BlockSize]
	target := make([]byte, len(byteText)-aes.BlockSize)
	copy(target, byteText[aes.BlockSize:])
	decryptStream := cipher.NewCTR(block, nonce)
	decryptStream.XORKeyStream(target, target)
	return string(target), nil
}

func ShuffleSlice[S any](s []S) []S {
	if len(s) <= 1 {
		return s
	}
	cs := slices.Clone(s)
	mrand.Shuffle(len(cs), func(i, j int) { cs[i], cs[j] = cs[j], cs[i] })
	return cs
}
