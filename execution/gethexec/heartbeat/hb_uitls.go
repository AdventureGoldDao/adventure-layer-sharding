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
	addField := func(value string) {
		builder.WriteString(fmt.Sprintf("%d", len(value)))
		builder.WriteString(value)
	}
	addField(contractAddress)
	addField(accountPublicKey)
	addField(strconv.Itoa(interval))
	addField(strconv.FormatInt(unx, 10))
	addField(key)
	if start {
		addField("start")
	} else {
		addField("stop")
	}
	h.Write([]byte(builder.String()))
	return hex.EncodeToString(h.Sum(nil)), nil
}
