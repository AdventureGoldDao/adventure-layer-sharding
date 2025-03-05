package heartbeat

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"strconv"
	"strings"
	"time"
)

func generateSignature(contractAddress, accountPublicKey string, interval int, start bool, unx int64) (string, error) {
	key := os.Getenv("HEART_BEAT_SIGN_KEY")
	if key == "" {
		log.Error("HEART_BEAT_SIGN_KEY is not set")
		return "", fmt.Errorf("HEART_BEAT_SIGN_KEY is not set")
	}
	if unx < time.Now().Unix()-10 || unx > time.Now().Unix()+10 {
		return "", fmt.Errorf("invalid timestamp")
	}
	h := sha256.New()
	var builder strings.Builder
	builder.WriteString(contractAddress)
	builder.WriteString(accountPublicKey)
	builder.WriteString(strconv.Itoa(interval))
	builder.WriteString(strconv.FormatInt(unx, 10))
	builder.WriteString(key)
	if start {
		builder.WriteString("start")
	} else {
		builder.WriteString("stop")
	}
	h.Write([]byte(builder.String()))
	return hex.EncodeToString(h.Sum(nil)), nil
}
