package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

// -------------------------------
// 🔗 Connect to Aptos Node
// -------------------------------
func ConnectClient(rpcURL string) *aptos.Client {
	client, err := aptos.NewClient(rpcURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Aptos network: %v", err)
	}
	return client
}

// -------------------------------
// 🧬 Create a New Account
// -------------------------------
func CreateAccount() (privateKeyHex string, address aptos.AccountAddress) {
	// Generate ED25519 key pair
	account, err := aptos.GenerateKeys()
	if err != nil {
		log.Fatalf("❌ Failed to generate private key: %v", err)
	}

	privateKeyHex = hex.EncodeToString(account.PrivateKey.Seed())
	address = account.Address

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key (Hex):", privateKeyHex)
	fmt.Println("🏦 Address:", address.String())

	return privateKeyHex, address
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func LoadAccount(privateKeyHex string) (*aptos.Account, aptos.AccountAddress) {
	// Decode hex private key
	seed, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}

	// Create account from private key
	privateKey := ed25519.NewKeyFromSeed(seed)
	account, err := aptos.NewAccountFromPrivateKey(privateKey)
	if err != nil {
		log.Fatalf("❌ Failed to load account: %v", err)
	}
	return account, account.Address
}

// -------------------------------
// 💰 Get Account Balance
// -------------------------------
func GetBalance(client *aptos.Client, address aptos.AccountAddress) float64 {
	ctx := context.Background()
	resourceType := "0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>"
	resource, err := client.AccountResource(ctx, address.String(), resourceType)
	if err != nil {
		// If account or resource doesn't exist, return 0 balance
		return 0.0
	}

	// Parse balance from resource
	data, ok := resource.(map[string]interface{})
	if !ok {
		log.Fatalf("❌ Failed to parse resource for %s", address.String())
	}
	coin, ok := data["data"].(map[string]interface{})["coin"].(map[string]interface{})
	if !ok {
		log.Fatalf("❌ Failed to parse coin data for %s", address.String())
	}
	balanceStr, ok := coin["value"].(string)
	if !ok {
		log.Fatalf("❌ Failed to parse balance value for %s", address.String())
	}
	balanceInt, err := strconv.ParseUint(balanceStr, 10, 64)
	if err != nil {
		log.Fatalf("❌ Failed to parse balance: %v", err)
	}
	return float64(balanceInt) / 1e8 // Convert Octas to APT
}

// -------------------------------
// 🚀 Send Transaction
// -------------------------------
func SendTransaction(client *aptos.Client, account *aptos.Account, toAddress aptos.AccountAddress, amountAPT float64) {
	ctx := context.Background()
	amountOctas := uint64(amountAPT * 1e8)

	// Build payload: aptos_coin::transfer
	payload := &aptos.TransactionPayload{
		Type: "entry_function_payload",
		Function: aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: aptos.AccountOne,
				Name:    "aptos_coin",
			},
			Name:      "transfer",
			TypeArgs:  []string{},
			Args:      []interface{}{toAddress.String(), amountOctas},
		},
	}

	// Build, sign, and submit transaction
	hash, err := client.BuildSignAndSubmitTransaction(ctx, account, payload)
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Hash: %s\n", hash)
}

// -------------------------------
// ⚙️ Utility Conversions
// -------------------------------
func OctasToAPT(octas uint64) float64 {
	return float64(octas) / 1e8
}

func APTToOctas(apt float64) uint64 {
	return uint64(apt * 1e8)
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// -------------------------------
	// 🌐 Aptos Networks
	// -------------------------------
	networks := map[string]string{
		"Aptos Mainnet": "https://fullnode.mainnet.aptoslabs.com/v1",
		"Aptos Testnet": "https://fullnode.testnet.aptoslabs.com/v1",
	}

	// 1️⃣ Create a new account (or load existing)
	privateKeyHex, address := CreateAccount()
	// To load: account, address := LoadAccount("YOUR_PRIVATE_KEY_HEX")

	fmt.Println("\n🏦 Wallet Address:", address.String())
	fmt.Println("\n🔑 Private Key (Hex):", privateKeyHex)

	// 2️⃣ Check balances on Mainnet and Testnet
	fmt.Println("\n💰 Balances:")
	for name, rpc := range networks {
		client := ConnectClient(rpc)
		balance := GetBalance(client, address)
		fmt.Printf("%s: %.6f APT\n", name, balance)
	}

	// 3️⃣ Example: Send Transaction (Uncomment to use)
	/*
		client := ConnectClient(networks["Aptos Testnet"])
		account, _ := LoadAccount(privateKeyHex)
		toAddress, err := aptos.AccountAddressFromHex("0xRECIPIENT_ADDRESS_HERE") // Replace with valid address
		if err != nil {
			log.Fatalf("❌ Invalid recipient address: %v", err)
		}
		SendTransaction(client, account, toAddress, 0.1) // Send 0.1 APT
	*/
}