package heartbeat

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

func (hb *HeartBeatAPI) sendHeartBeatTransaction(ctx context.Context, task *ContractTask, key *ecdsa.PrivateKey, data []byte) error {
	task.SendTxMutex.Lock()
	defer task.SendTxMutex.Unlock()

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)

	gasLimit, err := hb.estimateGas(ctx, &fromAddr, &task.ContractAddress, data)
	if err != nil {
		log.Error("Failed to estimate gas", "err", err)
		return nil
	}

	gasPrice, err := hb.gasPrice(ctx)
	if err != nil {
		log.Error("Failed to fetch gas price", "err", err)
		return nil
	}

	nonce, err := hb.b.APIBackend().GetPoolNonce(ctx, fromAddr)
	if err != nil {
		log.Error("Failed to fetch nonce", "err", err)
		return nil
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &task.ContractAddress,
		Value:    new(big.Int),
		Gas:      gasLimit,
		GasPrice: big.NewInt(gasPrice.ToInt().Int64() * defaultGasMultiplier),
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
	if err != nil {
		log.Error("Failed to sign tx", "err", err)
		return nil
	}
	if err := checkTxFee(tx.GasPrice(), tx.Gas(), hb.b.APIBackend().RPCTxFeeCap()); err != nil {
		return fmt.Errorf("tx fee estimation failed: %w", err)
	}
	if err := hb.b.APIBackend().SendTx(ctx, signedTx); err != nil {
		log.Error("Failed to send tx", "err", err)
		return nil
	}
	//go func() {
	//	var receipt *types.Receipt
	//	for {
	//		time.Sleep(2 * time.Second)
	//		receipts, err := hb.b.GetReceipts(ctx, signedTx.Hash())
	//		log.Info("GetReceipts", "receipts", receipts)
	//		if err != nil {
	//			log.Error("Failed to get receipts: %v", err)
	//		}
	//		if len(receipts) > 0 {
	//			receipt = receipts[0]
	//			break
	//		}
	//	}
	//
	//	actualGasFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), signedTx.GasPrice())
	//	log.Info("sendHeartBeatTransaction",
	//		"hash", signedTx.Hash().Hex(),
	//		"ContractAddress", task.ContractAddress.Hex(),
	//		"gas", signedTx.Gas(),
	//		"actualGasFee", actualGasFee.String(),
	//	)
	//}()
	log.Info("sendHeartBeatTransaction",
		"hash", signedTx.Hash().Hex(),
		"ContractAddress", task.ContractAddress.Hex(),
		"gas", signedTx.Gas(),
	)

	return nil
}
