package encoder

import (
	"errors"
	"math/big"

	"github.com/google/uuid"
)

const base62 = `0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

var (
	ErrInvalidBase62 = errors.New("invalid base62 character")
	ErrInvalidId     = errors.New("invalid id")
)

func EncodeUUIDToString(id uuid.UUID) (string, error) {
	if id == uuid.Nil {
		return "", ErrInvalidId
	}
	var shortened []byte                       // контейнер для закодированной строки
	identifier := new(big.Int).SetBytes(id[:]) // приводим uuid к bigInteger, это делимое
	base := big.NewInt(62)                     // делитель
	mod := new(big.Int)                        // остаток
	zero := big.NewInt(0)                      // шоткат для нуля

	for identifier.Cmp(zero) > 0 {
		identifier.DivMod(identifier, base, mod)
		shortened = append(shortened, base62[mod.Int64()])
	}
	for i, j := 0, len(shortened)-1; i < j; i, j = i+1, j-1 {
		shortened[i], shortened[j] = shortened[j], shortened[i]
	}
	return string(shortened), nil
}
