package humaycommon

import (
	"crypto/sha256"
)

func Hash256(body []byte, key string) [sha256.Size]byte {
	data := append([]byte(key), body...)

	return sha256.Sum256(data)
}
