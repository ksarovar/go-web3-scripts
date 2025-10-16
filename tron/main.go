package main

import (
	// "context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/client/transaction"
	// "github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	// "github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/mr-tron/base58"
	"google.golang.org/grpc"
)

// -------------------------------
// 🔗 Connect to RPC
// -------------------------------
func ConnectClient(rpcURL string) *client.GrpcClient {
	c := client.NewGrpcClient(rpcURL)
	err := c.Start(grpc.WithInsecure())
	if err != nil {
		log.Fatalf("❌ Failed to connect to Tron network: %v", err)
	}
	return c
}

// -------------------------------
// 🧬 Create a New Account
// -------------------------------
func CreateAccount() (privateKeyHex string, addr address.Address) {
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

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	hash := crypto.Keccak256(publicKeyBytes[1:])
	ethAddress := hash[12:]

	tronAddrBytes := append([]byte{0x41}, ethAddress...)
	h1 := sha256.Sum256(tronAddrBytes)
	h2 := sha256.Sum256(h1[:])
	checksum := h2[:4]
	tronAddrBytes = append(tronAddrBytes, checksum...)

	base58Addr := base58.Encode(tronAddrBytes)
	addr, err = address.Base58ToAddress(base58Addr)
	if err != nil {
		log.Fatalf("❌ Failed to create address: %v", err)
	}

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key:", privateKeyHex)
	fmt.Println("🏦 Address:", addr.String())

	return privateKeyHex, addr
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func LoadAccount(privateKeyHex string) (*ecdsa.PrivateKey, address.Address) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	hash := crypto.Keccak256(publicKeyBytes[1:])
	ethAddress := hash[12:]

	tronAddrBytes := append([]byte{0x41}, ethAddress...)
	h1 := sha256.Sum256(tronAddrBytes)
	h2 := sha256.Sum256(h1[:])
	checksum := h2[:4]
	tronAddrBytes = append(tronAddrBytes, checksum...)

	base58Addr := base58.Encode(tronAddrBytes)
	addr, err := address.Base58ToAddress(base58Addr)
	if err != nil {
		log.Fatalf("❌ Failed to create address: %v", err)
	}

	return privateKey, addr
}

// -------------------------------
// 💰 Get Account Balance
// -------------------------------
func GetBalance(c *client.GrpcClient, addr address.Address) *big.Float {
	account, err := c.GetAccount(addr.String())
	if err != nil && err.Error() != "account not found" {
		log.Fatalf("❌ Failed to get balance: %v", err)
	}
	balanceSun := int64(0)
	if account != nil {
		balanceSun = account.Balance
	}

	fbalance := new(big.Float).SetInt64(balanceSun)
	trxValue := new(big.Float).Quo(fbalance, big.NewFloat(1e6))
	return trxValue
}

// -------------------------------
// 🚀 Send Transaction
// -------------------------------
func SendTransaction(c *client.GrpcClient, privateKey *ecdsa.PrivateKey, toAddr address.Address, amountTrx float64) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddr := deriveTronAddress(publicKeyECDSA)

	amountSun := int64(amountTrx * 1e6)

	txExt, err := c.Transfer(fromAddr.String(), toAddr.String(), amountSun)
	if err != nil {
		log.Fatalf("❌ Failed to create transaction: %v", err)
	}

	signedTx, err := transaction.SignTransactionECDSA(txExt.Transaction, privateKey)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	result, err := c.Broadcast(signedTx)
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}
	if !result.Result {
		log.Fatalf("❌ Transaction failed: %s", result.Message)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Hash: %s\n", hex.EncodeToString(txExt.Txid))
}

// Helper function to derive Tron address
func deriveTronAddress(publicKeyECDSA *ecdsa.PublicKey) address.Address {
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	hash := crypto.Keccak256(publicKeyBytes[1:])
	ethAddress := hash[12:]

	tronAddrBytes := append([]byte{0x41}, ethAddress...)
	h1 := sha256.Sum256(tronAddrBytes)
	h2 := sha256.Sum256(h1[:])
	checksum := h2[:4]
	tronAddrBytes = append(tronAddrBytes, checksum...)

	base58Addr := base58.Encode(tronAddrBytes)
	addr, err := address.Base58ToAddress(base58Addr)
	if err != nil {
		log.Fatalf("❌ Failed to derive address: %v", err)
	}
	return addr
}

// -------------------------------
// ⚙️ Utility Conversions
// -------------------------------
func SunToTrx(sun int64) *big.Float {
	fsun := new(big.Float).SetInt64(sun)
	return new(big.Float).Quo(fsun, big.NewFloat(1e6))
}

func TrxToSun(trx float64) int64 {
	return int64(trx * 1e6)
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// -------------------------------
	// 🌐 Mainnet RPCs
	// -------------------------------
	mainnets := map[string]string{
		"Tron Mainnet": "grpc.trongrid.io:50051",
	}

	// -------------------------------
	// 🌐 Testnet RPCs
	// -------------------------------
	testnets := map[string]string{
		"Shasta Testnet": "grpc.shasta.trongrid.io:50051",
		"Nile Testnet":   "grpc.nile.trongrid.io:50051",
	}

	// 1️⃣ Create a new account (or load existing)
	privateKeyHex, addr := CreateAccount()
	// privateKey, addr := LoadAccount("YOUR_PRIVATE_KEY_HEX")
	fmt.Println("\n🏦 Wallet Address:", addr.String())
	fmt.Println("\n🏦 privateKeyHex:", privateKeyHex)

	// 2️⃣ Check balances on Mainnets
	fmt.Println("\n💰 Mainnet Balances:")
	for name, rpc := range mainnets {
		c := ConnectClient(rpc)
		defer c.Stop()
		balance := GetBalance(c, addr)
		fmt.Printf("%s: %f TRX\n", name, balance)
	}

	// 3️⃣ Check balances on Testnets
	fmt.Println("\n💰 Testnet Balances:")
	for name, rpc := range testnets {
		c := ConnectClient(rpc)
		defer c.Stop()
		balance := GetBalance(c, addr)
		fmt.Printf("%s: %f TRX\n", name, balance)
	}
}