package heartbeat

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/arbitrum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

// GasPrice returns a suggestion for a gas price for legacy transactions.
func (hb *HeartBeatAPI) gasPrice(ctx context.Context) (*hexutil.Big, error) {
	tipCap, err := hb.b.APIBackend().SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}
	if head := hb.b.APIBackend().CurrentHeader(); head.BaseFee != nil {
		tipCap.Add(tipCap, head.BaseFee)
	}
	return (*hexutil.Big)(tipCap), err
}

func (hb *HeartBeatAPI) estimateGas(ctx context.Context, fromAddr *common.Address, to *common.Address, data []byte) (uint64, error) {
	args := arbitrum.TransactionArgs{
		From:  fromAddr,
		To:    to,
		Data:  (*hexutil.Bytes)(&data),
		Value: (*hexutil.Big)(big.NewInt(0)),
	}
	blockNrOrHash := rpc.BlockNumberOrHashWithNumber(rpc.PendingBlockNumber)
	res, err := arbitrum.EstimateGas(ctx, hb.b.APIBackend(), args, blockNrOrHash, nil, hb.b.APIBackend().RPCGasCap())
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}
	return uint64(res), nil
}

func checkTxFee(gasPrice *big.Int, gas uint64, cap float64) error {
	// Short circuit if there is no cap for transaction fee at all.
	if cap == 0 {
		return nil
	}
	feeEth := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas))), new(big.Float).SetInt(big.NewInt(params.Ether)))
	feeFloat, _ := feeEth.Float64()
	if feeFloat > cap {
		return fmt.Errorf("tx fee (%.2f ether) exceeds the configured cap (%.2f ether)", feeFloat, cap)
	}
	return nil
}
