package encoder

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"
)

// EncodeURL генерирует короткий URL на основе оригинального URL и счетчика соли.
// В случае, если количество повторных попыток больше нуля,
// то строка мутируется, чтобы перегенерировать сокращение.
func EncodeURL(originalURL string, saltCounter int32) string {
	var input string
	if saltCounter > 0 {
		input = fmt.Sprintf("%s-%d", originalURL, saltCounter)
	} else {
		input = originalURL
	}
	hashBytes := md5.Sum([]byte(input)) //nolint:gosec
	hexString := hex.EncodeToString(hashBytes[:])
	return hexString[:7]
}
