package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

// ConnectStellarClient connects to a Stellar network
func ConnectStellarClient(networkURL string) *horizonclient.Client {
	client := horizonclient.DefaultTestNetClient
	if networkURL == "https://horizon.stellar.org" {
		client = horizonclient.DefaultPublicNetClient
	}
	return client
}

// CreateStellarAccount generates a new Stellar account
func CreateStellarAccount() (seed string, address string) {
	kp, err := keypair.Random()
	if err != nil {
		log.Fatalf("❌ Failed to generate keypair: %v", err)
	}

	seed = kp.Seed()
	address = kp.Address()

	fmt.Println("✅ New Stellar account created:")
	fmt.Println("🔑 Seed:", seed)
	fmt.Println("🏦 Address:", address)

	return seed, address
}

// LoadStellarAccount loads an existing Stellar account from a seed
func LoadStellarAccount(seed string) (*keypair.Full, error) {
	kp, err := keypair.ParseFull(seed)
	if err != nil {
		return nil, fmt.Errorf("❌ Invalid seed: %v", err)
	}
	return kp, nil
}

// GetStellarBalance retrieves the balance of a Stellar account
func GetStellarBalance(client *horizonclient.Client, address string) *big.Float {
	account, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: address})
	if err != nil {
		// Account not found or error, return nil
		return nil
	}

	var balanceXLM *big.Float
	for _, balance := range account.Balances {
		if balance.Asset.Type == "native" {
			balanceFloat, _ := new(big.Float).SetString(balance.Balance)
			balanceXLM = balanceFloat
			break
		}
	}
	return balanceXLM
}

// SendStellarTransaction sends a Stellar transaction
func SendStellarTransaction(client *horizonclient.Client, kp *keypair.Full, toAddress string, amountXLM string) {
	account, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: kp.Address()})
	if err != nil {
		log.Fatalf("❌ Failed to get account detail: %v", err)
	}

	paymentOp := txnbuild.Payment{
		Destination: toAddress,
		Amount:      amountXLM,
		Asset:       txnbuild.NativeAsset{},
	}

	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &account,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{&paymentOp},
			BaseFee:              txnbuild.MinBaseFee,
			Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
		},
	)
	if err != nil {
		log.Fatalf("❌ Failed to build transaction: %v", err)
	}

	tx, err = tx.Sign(network.TestNetworkPassphrase, kp)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	resp, err := client.SubmitTransaction(tx)
	if err != nil {
		log.Fatalf("❌ Failed to submit transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Hash: %s\n", resp.Hash)
}

// StroopsToXLM converts Stroops to XLM
func StroopsToXLM(stroops int64) *big.Float {
	return new(big.Float).Quo(big.NewFloat(float64(stroops)), big.NewFloat(1e7))
}

// XLMToStroops converts XLM to Stroops
func XLMToStroops(xlm float64) int64 {
	return int64(xlm * 1e7)
}

func main() {
	// Stellar network configurations
	stellarNetworks := map[string]string{
		"Stellar Mainnet": "https://horizon.stellar.org",
		"Stellar Testnet": "https://horizon-testnet.stellar.org",
	}

	// Create a new Stellar account
	seed, address := CreateStellarAccount()
	fmt.Println("\n🏦 Stellar Wallet Address:", address)
	fmt.Println("\n🏦 Stellar Seed:", seed)

	// Check balances on Stellar networks
	fmt.Println("\n💰 Stellar Balances:")
	for name, url := range stellarNetworks {
		client := ConnectStellarClient(url)
		balance := GetStellarBalance(client, address)
		if balance != nil {
			fmt.Printf("%s: %f XLM\n", name, balance)
		} else {
			fmt.Printf("%s: Account not found or zero balance\n", name)
		}
	}
}