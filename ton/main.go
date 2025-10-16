package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	// "time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	// "github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

// -------------------------------
// 🔗 Connect to TON RPC
// -------------------------------
func ConnectClient(configURL string) ton.APIClientWrapped {
	client := liteclient.NewConnectionPool()

	// Add connections from the TON network config
	err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to TON network: %v", err)
	}

	// Create API client with retry support
	api := ton.NewAPIClient(client).WithRetry()
	return api
}

// -------------------------------
// 🧬 Create a New TON Account
// -------------------------------
func CreateAccount(api ton.APIClientWrapped) ([]string, *address.Address) {
	// Generate a new mnemonic seed phrase
	seed := wallet.NewSeed()

	// Create a wallet (v4r2, workchain 0)
	w, err := wallet.FromSeed(api, seed, wallet.V4R2)
	if err != nil {
		log.Fatalf("❌ Failed to create wallet: %v", err)
	}

	addr := w.WalletAddress()
	fmt.Println("✅ New TON account created:")
	fmt.Println("🔑 Seed Phrase:", seed)
	fmt.Println("🏦 Address:", addr.String())
	return seed, addr
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func LoadAccount(api ton.APIClientWrapped, seed []string) (*wallet.Wallet, *address.Address) {
	// Create wallet from existing seed phrase
	w, err := wallet.FromSeed(api, seed, wallet.V4R2)
	if err != nil {
		log.Fatalf("❌ Failed to load wallet: %v", err)
	}

	addr := w.WalletAddress()
	return w, addr
}

// -------------------------------
// 💰 Get TON Account Balance
// -------------------------------
func GetBalance(api ton.APIClientWrapped, addr *address.Address) *big.Int {
	// Get the latest block ID
	master, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Printf("⚠️ Failed to get masterchain info: %v (balance 0)", err)
		return big.NewInt(0)
	}

	// Fetch account state
	account, err := api.GetAccount(context.Background(), master, addr)
	if err != nil || account == nil {
		log.Printf("⚠️ Account %s not found or inactive: %v (balance 0)", addr.String(), err)
		return big.NewInt(0)
	}

	if !account.IsActive {
		log.Printf("⚠️ Account %s is not active (balance 0)", addr.String())
		return big.NewInt(0)
	}

	return account.State.Balance.Nano()
}

// -------------------------------
// 🚀 Send TON Transaction
// -------------------------------
// func SendTransaction(api ton.APIClientWrapped, w *wallet.Wallet, toAddr *address.Address, amountTON float64) {
// 	// Convert TON to NanoTON
// 	amountBig := new(big.Int).SetInt64(int64(amountTON * 1e9))
// 	amount := tlb.FromNanoTON(amountBig)

// 	// Create transfer message
// 	transfer, err := w.BuildTransfer(toAddr, amount, true, "Sending TON")
// 	if err != nil {
// 		log.Fatalf("❌ Failed to build transfer: %v", err)
// 	}

// 	// Send transaction with context timeout
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
// 	err = w.Send(ctx, api, transfer)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to send transaction: %v", err)
// 	}

// 	fmt.Printf("✅ Transaction sent successfully!\n🔗 To Address: %s\n🔗 Amount: %f TON\n", toAddr.String(), amountTON)
// }

// -------------------------------
// ⚙️ Utility Conversions
// -------------------------------
func NanoTONToTON(nano *big.Int) float64 {
	if nano == nil {
		return 0.0
	}
	f := new(big.Float).SetInt(nano)
	f = f.Quo(f, big.NewFloat(1e9))
	result, _ := f.Float64()
	return result
}

func TONToNanoTON(ton float64) *big.Int {
	return new(big.Int).SetInt64(int64(ton * 1e9))
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// -------------------------------
	// 🌐 Mainnet Configs
	// -------------------------------
	mainnets := map[string]string{
		"TON Mainnet": "https://ton-blockchain.github.io/global.config.json",
	}

	// -------------------------------
	// 🌐 Testnet Configs
	// -------------------------------
	testnets := map[string]string{
		"TON Testnet": "https://ton-blockchain.github.io/testnet-global.config.json",
	}

	// Connect to testnet to create/load wallet
	testnetAPI := ConnectClient("https://ton-blockchain.github.io/testnet-global.config.json")

	// 1️⃣ Create a new account (or load existing)
	seed, addr := CreateAccount(testnetAPI)
	// To load an existing account:
	// w, addr := LoadAccount(testnetAPI, []string{"your", "seed", "phrase", "here"})
	w, _ := LoadAccount(testnetAPI, seed)
	fmt.Println("\n🏦 w :", w)

	fmt.Println("\n🏦 Wallet Address:", addr.String())
	fmt.Println("🔑 Seed Phrase:", seed)

	// 2️⃣ Check balances on Mainnets
	fmt.Println("\n💰 Mainnet Balances:")
	for name, config := range mainnets {
		api := ConnectClient(config)
		balance := GetBalance(api, addr)
		fmt.Printf("%s: %f TON\n", name, NanoTONToTON(balance))
	}

	// 3️⃣ Check balances on Testnets
	fmt.Println("\n💰 Testnet Balances:")
	for name, config := range testnets {
		api := ConnectClient(config)
		balance := GetBalance(api, addr)
		fmt.Printf("%s: %f TON\n", name, NanoTONToTON(balance))
	}

	// 4️⃣ Example: Send 0.01 TON to another address (Testnet)
	// ⚠️ Ensure the account has enough TON on the testnet
	recipient := "EQCD39VS5jcptHL8vMjEXrzGaRcCVYto7HUn4bpAOg8xqB2N" // TON Foundation address (replace with a valid testnet address)

	toAddr, err := address.ParseAddr(recipient)
	fmt.Println("\n🏦 toAddr Address:", toAddr)

	if err != nil {
		log.Fatalf("❌ Invalid recipient address: %v", err)
	}

	// SendTransaction(testnetAPI, w, toAddr, 0.01)
}