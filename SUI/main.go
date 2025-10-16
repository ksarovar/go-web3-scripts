package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/coming-chat/go-sui/v2/account"
	"github.com/coming-chat/go-sui/v2/client"
	sui_types "github.com/coming-chat/go-sui/v2/types"
)

// Connect to RPC
func ConnectClient(rpcURL string) *client.Client {
	cli, err := client.Dial(rpcURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Sui network: %v", err)
	}
	return cli
}

// Create a new account
func CreateAccount() (privateKeyHex string, addressStr string) {
	// Correct way: create SignatureScheme first
	scheme, err := sui_types.NewSignatureScheme(0) // 0 = Ed25519
	if err != nil {
		log.Fatalf("❌ Failed to create SignatureScheme: %v", err)
	}

	acc := account.NewAccount(scheme, nil)

	privateKeyHex = hex.EncodeToString(acc.KeyPair.PrivateKey())
	addressStr = acc.Address

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key:", privateKeyHex)
	fmt.Println("🏦 Address:", addressStr)

	return privateKeyHex, addressStr
}

// Load existing account
func LoadAccount(privateKeyHex string) *account.Account {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}

	scheme, err := sui_types.NewSignatureScheme(0)
	if err != nil {
		log.Fatalf("❌ Failed to create SignatureScheme: %v", err)
	}

	acc := account.NewAccount(scheme, privBytes)
	return acc
}

// Get balance
func GetBalance(cli *client.Client, address string) string {
	balance, err := cli.GetBalance(context.Background(), address, sui_types.SUI_COIN_TYPE)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get balance: %v", err)
		return "0"
	}
	return balance.TotalBalance.String()
}

// Main
func main() {
	mainnet := "https://fullnode.mainnet.sui.io:443"
	testnet := "https://fullnode.testnet.sui.io:443"

	privateKeyHex, address := CreateAccount()

	fmt.Println("\n🏦 Wallet Address:", address)
	fmt.Println("🔑 Private Key:", privateKeyHex)

	fmt.Println("\n💰 Checking Mainnet Balance...")
	mainCli := ConnectClient(mainnet)
	fmt.Println("Mainnet Balance:", GetBalance(mainCli, address), "SUI")

	fmt.Println("\n💰 Checking Testnet Balance...")
	testCli := ConnectClient(testnet)
	fmt.Println("Testnet Balance:", GetBalance(testCli, address), "SUI")
}
