package heartbeat

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/arbitrum"
	"github.com/ethereum/go-ethereum/log"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func NewHeartBeatAPI(b *arbitrum.Backend) *HeartBeatAPI {
	hb := &HeartBeatAPI{b}
	stateManager = NewStateManager()
	if stateManager != nil {
		stateManager.ContractMap.Range(func(key, value interface{}) bool {
			task, ok := value.(*ContractTask)
			if !ok {
				log.Error("Failed to assert value to ContractTask")
				return false
			}
			ctx, cancel := context.WithCancel(context.Background())
			task.CancelFunc = cancel
			go hb.startPolling(ctx, task)
			stateManager.Count++
			return true
		})
	}
	return hb
}

func (hb *HeartBeatAPI) ManageContractTask(contractAddress, accountPublicKey string, interval int, start bool, signature string) string {
	if contractAddress == "" || accountPublicKey == "" || interval <= 100 {
		return "params err!"
	}
	expectedSignature := generateSignature(contractAddress, accountPublicKey, interval)
	if expectedSignature == "" || expectedSignature != signature {
		return "invalid signature"
	}

	accountAddr := common.HexToAddress(accountPublicKey)
	contractAddr := common.HexToAddress(contractAddress)
	if start {
		return hb.startTask(accountAddr, contractAddr, interval)
	}
	return hb.stopTask(accountAddr)
}

func (hb *HeartBeatAPI) GetActiveHeartBeats() []map[string]string {
	var activeHeartBeats []map[string]string
	stateManager.ContractMap.Range(func(key, value interface{}) bool {
		task, ok := value.(*ContractTask)
		if !ok {
			return false
		}
		activeHeartBeats = append(activeHeartBeats, map[string]string{
			"contractAddress":  task.ContractAddress.Hex(),
			"accountPublicKey": task.AccountPublicKey.Hex(),
		})
		return true
	})
	return activeHeartBeats
}

func (hb *HeartBeatAPI) startTask(accountPublicKey, contractAddress common.Address, interval int) string {
	stateManagerMutex.Lock()
	defer stateManagerMutex.Unlock()

	if stateManager.LimitStatus() {
		return "heartbeat task limit ..."
	}
	if task := stateManager.LoadOne(accountPublicKey); task != nil {
		task.CancelFunc()
	}
	ctx, cancel := context.WithCancel(context.Background())
	task := &ContractTask{
		CancelFunc:       cancel,
		Interval:         time.Duration(interval) * time.Millisecond,
		AccountPublicKey: accountPublicKey,
		ContractAddress:  contractAddress,
	}

	go hb.startPolling(ctx, task)

	stateManager.Save(task)
	return fmt.Sprintf("Started polling for contract: %s", contractAddress.Hex())
}

func (hb *HeartBeatAPI) stopTask(addr common.Address) string {
	task := stateManager.LoadOne(addr)
	if task == nil {
		return fmt.Sprintf("No active task found for accountAddr: %s", addr.Hex())
	}
	task.CancelFunc()
	stateManager.Delete(addr)
	log.Info("Stopped polling for contract: ", task.ContractAddress.Hex())
	return fmt.Sprintf("Stop polling for contract: %s", addr.Hex())
}
