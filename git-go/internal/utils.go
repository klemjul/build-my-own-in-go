package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func CreateSha1Hex(content []byte) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write(content)
	if err != nil {
		return "", fmt.Errorf("CreateHash: failed to hash, %v", err)
	}
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)
	return hashHex, nil
}
