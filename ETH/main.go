package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// -------------------------------
// 🔗 Connect to RPC
// -------------------------------
func ConnectClient(rpcURL string) *ethclient.Client {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Ethereum network: %v", err)
	}
	return client
}

// -------------------------------
// 🧬 Create a New Account
// -------------------------------
func CreateAccount() (privateKeyHex string, address common.Address) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		log.Fatalf("❌ Failed to generate private key: %v", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex = hex.EncodeToString(privateKeyBytes)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address = crypto.PubkeyToAddress(*publicKeyECDSA)

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key:", privateKeyHex)
	fmt.Println("🏦 Address:", address.Hex())

	return privateKeyHex, address
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func LoadAccount(privateKeyHex string) (*ecdsa.PrivateKey, common.Address) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, address
}

// -------------------------------
// 💰 Get Account Balance
// -------------------------------
func GetBalance(client *ethclient.Client, address common.Address) *big.Float {
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatalf("❌ Failed to get balance: %v", err)
	}

	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(1e18))
	return ethValue
}

// -------------------------------
// 🚀 Send Transaction
// -------------------------------
func SendTransaction(client *ethclient.Client, privateKey *ecdsa.PrivateKey, toAddress common.Address, amountEther float64) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("❌ Failed to get nonce: %v", err)
	}

	value := new(big.Int)
	value.SetString(fmt.Sprintf("%.0f", amountEther*1e18), 10)

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to suggest gas price: %v", err)
	}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to get chain ID: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Hash: %s\n", signedTx.Hash().Hex())
}

// -------------------------------
// ⚙️ Utility Conversions
// -------------------------------
func WeiToEther(wei *big.Int) *big.Float {
	fwei := new(big.Float).SetInt(wei)
	return new(big.Float).Quo(fwei, big.NewFloat(1e18))
}

func EtherToWei(eth float64) *big.Int {
	value := new(big.Int)
	value.SetString(fmt.Sprintf("%.0f", eth*1e18), 10)
	return value
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// -------------------------------
	// 🌐 Mainnet RPCs
	// -------------------------------
	mainnets := map[string]string{
		"Ethereum Mainnet":     "https://eth.drpc.org",
		"Polygon Mainnet":      "https://polygon-bor-rpc.publicnode.com",
		"BNB Smart Chain":      "https://bsc-dataseed.binance.org/",
		"Arbitrum One":         "https://arb1.arbitrum.io/rpc",
		"Optimism":             "https://mainnet.optimism.io",
		"Ethereum Classic":     "https://ethereum-classic-mainnet.gateway.tatum.io",
		"Base":                 "https://base-mainnet.public.blastapi.io",
		"Linea":                "https://rpc.linea.build",
		"Scroll":               "https://rpc.scroll.io",
		"zkSync Era":           "https://mainnet.era.zksync.io",
		"Polygon zkEVM":        "https://zkevm-rpc.com",
		"ETHW":                 "https://mainnet.ethereumpow.org",
		"opBNB":                "https://opbnb.rpc.grove.city/v1/01fdb492",

	}

	// -------------------------------
	// 🌐 Testnet RPCs
	// -------------------------------
	testnets := map[string]string{
		"Base Sepolia Testnet":             "https://sepolia.base.org",
		"Polygon amoy Testnet":      "https://polygon-amoy.drpc.org",
		"BNB Smart Chain Testnet": "https://data-seed-prebsc-2-s3.bnbchain.org:8545",
		"Eth Sepolia Testnet":             "https://11155111.rpc.thirdweb.com",

	}

	// 1️⃣ Create a new account (or load existing)
	privateKeyHex, address := CreateAccount()
	// privateKey, address := LoadAccount("YOUR_PRIVATE_KEY_HEX")
	fmt.Println("\n🏦 Wallet Address:", address.Hex())
	fmt.Println("\n🏦 privateKeyHex:", privateKeyHex)


	// 2️⃣ Check balances on Mainnets
	fmt.Println("\n💰 Mainnet Balances:")
	for name, rpc := range mainnets {
		client := ConnectClient(rpc)
		balance := GetBalance(client, address)
		fmt.Printf("%s: %f ETH\n", name, balance)
	}

	// 3️⃣ Check balances on Testnets
	fmt.Println("\n💰 Testnet Balances:")
	for name, rpc := range testnets {
		client := ConnectClient(rpc)
		balance := GetBalance(client, address)
		fmt.Printf("%s: %f ETH\n", name, balance)
	}
}
