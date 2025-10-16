package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"

	// "github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

// -------------------------------
// Stacks Account and Transaction Structures
// -------------------------------
type StacksAccount struct {
	PrivateKey string
	Address    string
	Mnemonic   string
}

type StacksBalanceResponse struct {
	Balance string `json:"balance"`
}

type StacksBroadcastResponse struct {
	TxID  string `json:"txid"`
	Error string `json:"error"`
}

// -------------------------------
// 🔗 Connect to Stacks API
// -------------------------------
func connectStacksAPI(isMainnet bool) string {
	if isMainnet {
		return "https://api.mainnet.hiro.so"
	}
	return "https://api.testnet.hiro.so"
}

// -------------------------------
// 🧬 Create a New Account
// -------------------------------
func createStacksAccount(isMainnet bool) StacksAccount {
	// Generate a random 32-byte seed
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		log.Fatalf("❌ Failed to generate entropy: %v", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		log.Fatalf("❌ Failed to generate mnemonic: %v", err)
	}

	seed := bip39.NewSeed(mnemonic, "")
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		log.Fatalf("❌ Failed to generate master key: %v", err)
	}

	// Derive a key for Stacks (m/44'/5757'/0'/0/0)
	path := []uint32{44 + 0x80000000, 5757 + 0x80000000, 0 + 0x80000000, 0, 0}
	key := masterKey
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			log.Fatalf("❌ Failed to derive key: %v", err)
		}
	}

	privateKey := key.Key
	// Derive public key using btcec
	privKey, pubKey := btcec.PrivKeyFromBytes(privateKey)
	fmt.Println("🔑 Private Key:", privKey)

	publicKeyBytes := pubKey.SerializeCompressed()

	// Stacks address derivation (simplified C32 encoding)
	// Stacks uses version bytes (26 for mainnet, 21 for testnet) and RIPEMD160(SHA256(pubkey))
	hash := hash160(publicKeyBytes)
	var version byte = 26 // Mainnet
	if !isMainnet {
		version = 21 // Testnet
	}
	address := encodeC32Address(version, hash)

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key:", hex.EncodeToString(privateKey))
	fmt.Println("🏦 Address:", address)
	fmt.Println("📝 Mnemonic:", mnemonic)

	return StacksAccount{
		PrivateKey: hex.EncodeToString(privateKey),
		Address:    address,
		Mnemonic:   mnemonic,
	}
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func loadStacksAccount(privateKeyHex string, isMainnet bool) StacksAccount {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}

	// Derive public key using btcec
	_, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)
	publicKeyBytes := pubKey.SerializeCompressed()

	// Stacks address derivation
	hash := hash160(publicKeyBytes)
	var version byte = 26 // Mainnet
	if !isMainnet {
		version = 21 // Testnet
	}
	address := encodeC32Address(version, hash)

	return StacksAccount{
		PrivateKey: privateKeyHex,
		Address:    address,
	}
}

// -------------------------------
// 💰 Get Account Balance
// -------------------------------
func getStacksBalance(apiURL, address string) float64 {
	resp, err := http.Get(fmt.Sprintf("%s/v2/accounts/%s", apiURL, address))
	if err != nil {
		log.Printf("❌ Failed to get balance for %s: %v", address, err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read balance response: %v", err)
		return 0
	}

	var balanceResp StacksBalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		// log.Printf("❌ Failed to parse balance response: %v", err)
		return 0
	}

	balance, ok := new(big.Int).SetString(balanceResp.Balance, 16) // Balance in microSTX
	if !ok {
		log.Printf("❌ Failed to parse balance: %s", balanceResp.Balance)
		return 0
	}
	fbalance := new(big.Float).SetInt(balance)
	stxValue := new(big.Float).Quo(fbalance, big.NewFloat(1e6)) // Convert to STX
	stxFloat, _ := stxValue.Float64()
	return stxFloat
}

// -------------------------------
// 🚀 Send Transaction (STX Transfer)
// -------------------------------
func sendStacksTransaction(apiURL, privateKey, toAddress string, amountSTX float64) {
	// Placeholder: Stacks transactions require Clarity-based construction
	amountMicroSTX := new(big.Int).SetInt64(int64(amountSTX * 1e6)) // Convert to microSTX
	tx := map[string]interface{}{
		"recipient": toAddress,
		"amount":    amountMicroSTX.String(),
		"nonce":     "0",   // Simplified; fetch nonce from API
		"fee":       "180", // Fixed fee; use estimate API in production
	}

	txBytes, err := json.Marshal(tx)
	if err != nil {
		log.Fatalf("❌ Failed to marshal transaction: %v", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/v2/transactions", apiURL), "application/json", bytes.NewBuffer(txBytes))
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read transaction response: %v", err)
	}

	var broadcastResp StacksBroadcastResponse
	if err := json.Unmarshal(body, &broadcastResp); err != nil {
		log.Fatalf("❌ Failed to parse transaction response: %v", err)
	}

	if broadcastResp.Error != "" {
		log.Fatalf("❌ Transaction failed: %s", broadcastResp.Error)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 TxID: %s\n", broadcastResp.TxID)
}

// -------------------------------
// 🛠️ Utility: Compute RIPEMD160(SHA256(data))
// -------------------------------
func hash160(data []byte) []byte {
	sha256Hash := sha256.Sum256(data)
	ripemd160Hash := ripemd160.New()
	ripemd160Hash.Write(sha256Hash[:])
	return ripemd160Hash.Sum(nil)
}

// -------------------------------
// 🛠️ Utility: Encode C32 Address (Simplified)
// -------------------------------
func encodeC32Address(version byte, hash []byte) string {
	// Stacks uses C32 (base32 with custom alphabet) encoding
	// This is a simplified version; in production, use a proper C32 library
	data := append([]byte{version}, hash...)
	// For demo, return a placeholder address (real C32 encoding requires a library)
	prefix := "SP"
	if version == 21 {
		prefix = "ST"
	}
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(data)[:32]) // Simplified
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -----------------------
func main() {
	// Networks
	networks := map[string]bool{
		"Stacks Mainnet": true,
		"Stacks Testnet": false,
	}

	// 1️⃣ Create a new account (or load existing)
	account := createStacksAccount(false) // Testnet
	// account := loadStacksAccount("YOUR_PRIVATE_KEY_HEX", false)
	fmt.Println("\n🏦 Wallet Address:", account.Address)
	fmt.Println("\n🔑 Private Key:", account.PrivateKey)
	fmt.Println("\n📝 Mnemonic:", account.Mnemonic)

	// 2️⃣ Check balances
	fmt.Println("\n💰 Balances:")
	for name, isMainnet := range networks {
		apiURL := connectStacksAPI(isMainnet)
		balance := getStacksBalance(apiURL, account.Address)
		fmt.Printf("%s: %.6f STX\n", name, balance)
	}

	// 3️⃣ Example: Send 1 STX (uncomment to test)
	// toAddress := "ST..." // Replace with valid Stacks address
	// sendStacksTransaction(connectStacksAPI(false), account.PrivateKey, toAddress, 1)
}