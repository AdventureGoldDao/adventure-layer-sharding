package heartbeat

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func NewStateManager() *StateManager {
	ecdsa, err := crypto.HexToECDSA(os.Getenv("HEAT_BEAT_PRIVATE_KEY"))
	if err != nil {
		log.Error("Failed to load HEAT_BEAT_PRIVATE_KEY:", err)
		return nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error("unable to read users home directory: %w", err)
		return nil
	}

	dir := path.Join(homeDir, defaultStateDirname)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error("unable to create global configuration directory: %w", err)
		return nil
	}
	s := &StateManager{
		StateDir:   dir,
		PrivateKey: ecdsa,
	}
	err = s.LoadAll()
	if err != nil {
		log.Error("Failed to load NewStateManager:", err)
	}
	return s
}

func (sm *StateManager) LimitStatus() bool {
	hbLimitNum := os.Getenv("HEAT_BEAT_LIMIT_NUM")
	hbLimitNumInt, err := strconv.Atoi(hbLimitNum)
	if hbLimitNum == "" || err != nil {
		hbLimitNumInt = 8
	}
	return sm.Count >= hbLimitNumInt
}
func (sm *StateManager) Save(state *ContractTask) {
	sm.ContractMap.Store(state.AccountPublicKey.Hex(), state)
	sm.saveFile(state)
	sm.Count++
}

func (sm *StateManager) LoadOne(address common.Address) *ContractTask {
	if s, ok := sm.ContractMap.Load(address.Hex()); ok {
		return s.(*ContractTask)
	}
	return nil
}
func (sm *StateManager) LoadAll() error {
	entries, err := os.ReadDir(sm.StateDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) != 42 {
			continue
		}
		filePath := filepath.Join(sm.StateDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Error("Failed to read state file", "file", filePath, "error", err)
			continue
		}

		var state StateData
		if err := json.Unmarshal(data, &state); err != nil {
			log.Error("Failed to parse state file", "file", filePath, "error", err)
			continue
		}
		taskState := new(ContractTask)
		taskState.AccountPublicKey = common.HexToAddress(state.AccountPublicKey)
		taskState.ContractAddress = common.HexToAddress(state.ContractAddress)
		taskState.Interval = state.Interval
		sm.ContractMap.Store(state.AccountPublicKey, taskState)
	}
	return nil
}

func (sm *StateManager) Delete(address common.Address) {
	sm.ContractMap.Delete(address.Hex())
	sm.Count--
	file := filepath.Join(sm.StateDir, address.Hex())
	log.Info("heartbeat deleteFile", "file", file)
	err := os.Remove(file)
	if err != nil {
		log.Error("Error deleting file: %v\n", err)
	}
}

func (sm *StateManager) saveFile(state *ContractTask) {
	fileName := filepath.Join(sm.StateDir, state.AccountPublicKey.Hex())
	file, err := os.Create(fileName)
	if err != nil {
		log.Error("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	data := StateData{
		AccountPublicKey: state.AccountPublicKey.Hex(),
		ContractAddress:  state.ContractAddress.Hex(),
		Interval:         state.Interval,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Error marshaling JSON: %v\n", err)
		return
	}
	_, err = io.WriteString(file, string(jsonData))
	if err != nil {
		log.Error("Error writing to file: %v\n", err)
		return
	}
	log.Info("heartbeat saveFile success", "file", fileName)
}
