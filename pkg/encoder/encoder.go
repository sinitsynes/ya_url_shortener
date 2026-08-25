package encoder

import (
	"errors"
)

const base62 = `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

var ErrInvalidBase62 = errors.New("invalid base62 character")

func EncodeUrl(id int32) string {
	if id == 0 {
		return "0"
	}
	var shortened []byte
	for id > 0 {
		remainder := id % 62
		shortened = append(shortened, base62[remainder])
		id /= 62
	}
	for i, j := 0, len(shortened)-1; i < j; i, j = i+1, j-1 {
		shortened[i], shortened[j] = shortened[j], shortened[i]
	}
	return string(shortened)
}
