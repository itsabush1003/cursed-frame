package util

import (
	"crypto/rand"
	"encoding/base64"
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

func ShuffleSlice[S any](s []S) []S {
	if len(s) <= 1 {
		return s
	}
	cs := slices.Clone(s)
	mrand.Shuffle(len(cs), func(i, j int) { cs[i], cs[j] = cs[j], cs[i] })
	return cs
}
