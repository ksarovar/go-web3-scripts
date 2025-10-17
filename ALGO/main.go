package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
)

// ConnectClient connects to the Algorand network
func ConnectClient(algodAddress, algodToken string) *algod.Client {
	client, err := algod.MakeClient(algodAddress, algodToken)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Algorand network: %v", err)
	}
	return client
}

// CreateAccount generates a new Algorand account
func CreateAccount() (mnemonicPhrase string, address string) {
	account := crypto.GenerateAccount()
	mnemonicPhrase, err := mnemonic.FromPrivateKey(account.PrivateKey)
	if err != nil {
		log.Fatalf("❌ Failed to generate mnemonic: %v", err)
	}

	address = account.Address.String()

	fmt.Println("✅ New Algorand account created:")
	fmt.Println("🔑 Mnemonic:", mnemonicPhrase)
	fmt.Println("🏦 Address:", address)

	return mnemonicPhrase, address
}

// LoadAccount loads an existing Algorand account from a mnemonic
func LoadAccount(mnemonicPhrase string) (crypto.Account, error) {
	privateKey, err := mnemonic.ToPrivateKey(mnemonicPhrase)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("❌ Invalid mnemonic: %v", err)
	}

	account, err := crypto.AccountFromPrivateKey(privateKey)
	if err != nil {
		return crypto.Account{}, fmt.Errorf("❌ Failed to load account: %v", err)
	}

	return account, nil
}

// GetBalance retrieves the balance of an Algorand account
func GetBalance(client *algod.Client, address string) *big.Float {
	accountInfo, err := client.AccountInformation(address).Do(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get account info: %v", err)
	}

	balanceMicroalgos := accountInfo.Amount
	balanceAlgos := new(big.Float).Quo(big.NewFloat(float64(balanceMicroalgos)), big.NewFloat(1e6))
	return balanceAlgos
}

// SendTransaction sends an Algorand transaction
func SendTransaction(client *algod.Client, account crypto.Account, toAddress string, amountAlgos float64) {
	txParams, err := client.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get suggested params: %v", err)
	}

	amountMicroalgos := uint64(amountAlgos * 1e6)

	toAddr, err := types.DecodeAddress(toAddress)
	if err != nil {
		log.Fatalf("❌ Invalid to address: %v", err)
	}

	txn, err := transaction.MakePaymentTxn(account.Address.String(), toAddr.String(), uint64(txParams.Fee), amountMicroalgos, uint64(txParams.FirstRoundValid), uint64(txParams.LastRoundValid), nil, "", "", txParams.GenesisHash)
	if err != nil {
		log.Fatalf("❌ Failed to make transaction: %v", err)
	}

	txID, signedTxn, err := crypto.SignTransaction(account.PrivateKey, txn)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	sendResponse, err := client.SendRawTransaction(signedTxn).Do(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 TxID: %s\n", txID)
	fmt.Printf("🔗 Confirmed TxID: %s\n", sendResponse)
}

// MicroalgosToAlgos converts microalgos to Algos
func MicroalgosToAlgos(microalgos uint64) *big.Float {
	return new(big.Float).Quo(big.NewFloat(float64(microalgos)), big.NewFloat(1e6))
}

// AlgosToMicroalgos converts Algos to microalgos
func AlgosToMicroalgos(algos float64) uint64 {
	return uint64(algos * 1e6)
}

func main() {
	// Algorand network configurations
	algorandNetworks := map[string]map[string]string{
		"Algorand Mainnet": {
			"address": "https://mainnet-api.algonode.cloud",
			"token":   "",
		},
		"Algorand Testnet": {
			"address": "https://testnet-api.algonode.cloud",
			"token":   "",
		},
	}

	// Create a new Algorand account
	mnemonicPhrase, address := CreateAccount()
	fmt.Println("\n🏦 Algorand Wallet Address:", address)
	fmt.Println("\n🏦 Algorand Mnemonic Phrase:", mnemonicPhrase)

	// Check balances on Algorand networks
	fmt.Println("\n💰 Algorand Balances:")
	for name, config := range algorandNetworks {
		client := ConnectClient(config["address"], config["token"])
		balance := GetBalance(client, address)
		fmt.Printf("%s: %f ALGO\n", name, balance)
	}
}