package heartbeat

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"os"
)

func generateSignature(contractAddress, accountPublicKey string, interval int) string {
	key := os.Getenv("HEART_BEAT_SIGN_KEY")
	if key == "" {
		log.Error("HEART_BEAT_SIGN_KEY is not set")
		return ""
	}
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s%s%d%s", contractAddress, accountPublicKey, interval, key)))
	return hex.EncodeToString(h.Sum(nil))
}
