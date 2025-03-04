package heartbeat

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// 验签方法
func verifySignature(data, signature, key string) (bool, error) {
	privKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return false, err
	}
	pubKey := privKey.Public()
	pubKeyECDSA, ok := pubKey.(*crypto.ECDSAPublicKey)
	if !ok {
		return false, fmt.Errorf("error casting public key to ECDSA")
	}

	hash := crypto.Keccak256Hash([]byte(data))
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}
	sig = sig[:len(sig)-1] // remove recovery id
	recoveredPubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}
	recoveredPubKeyECDSA, ok := recoveredPubKey.(*crypto.ECDSAPublicKey)
	if !ok {
		return false, fmt.Errorf("error casting recovered public key to ECDSA")
	}

	return pubKeyECDSA.X.Cmp(recoveredPubKeyECDSA.X) == 0 && pubKeyECDSA.Y.Cmp(recoveredPubKeyECDSA.Y) == 0, nil
}

// ManageContractTask接口
func (h *HeartbeatAPI) ManageContractTask(w http.ResponseWriter, r *http.Request) {
	// 从请求中获取数据和签名
	data := r.FormValue("data")
	signature := r.FormValue("signature")

	// 从环境变量读取密钥
	key := os.Getenv("HEART_BEAT_SIGN_KEY")
	if key == "" {
		http.Error(w, "Missing HEART_BEAT_SIGN_KEY environment variable", http.StatusInternalServerError)
		return
	}

	// 验签
	isValid, err := verifySignature(data, signature, key)
	if err != nil {
		http.Error(w, "Signature verification failed", http.StatusUnauthorized)
		return
	}
	if !isValid {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// 处理任务逻辑
	// ...
}
