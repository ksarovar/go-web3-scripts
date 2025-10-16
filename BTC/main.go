package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"github.com/btcsuite/btcd/btcec/v2" // Updated import
)

// -------------------------------
// Bitcoin Account and Transaction Structures
// -------------------------------
type BitcoinAccount struct {
	PrivateKey string
	Address    string
	WIF        string
}

type BitcoinUTXOResponse struct {
	TxID  string `json:"txid"`
	Vout  uint32 `json:"vout"`
	Value int64  `json:"value"`
}

// -------------------------------
// 🔗 Connect to Bitcoin API
// -------------------------------
func connectBitcoinAPI(isMainnet bool) string {
	if isMainnet {
		return "https://blockstream.info/api"
	}
	return "https://blockstream.info/testnet/api"
}

// -------------------------------
// 🧬 Create a New Account
// -------------------------------
func createBitcoinAccount(isMainnet bool) BitcoinAccount {
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

	// Derive a key for Bitcoin (m/44'/0'/0'/0/0 for mainnet, m/44'/1'/0'/0/0 for testnet)
	coinType := uint32(0)
	if !isMainnet {
		coinType = 1
	}
	path := []uint32{44 + 0x80000000, coinType + 0x80000000, 0 + 0x80000000, 0, 0}
	key := masterKey
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			log.Fatalf("❌ Failed to derive key: %v", err)
		}
	}

	privateKey := key.Key
	network := &chaincfg.MainNetParams
	if !isMainnet {
		network = &chaincfg.TestNet3Params
	}

	// Convert privateKey to *btcec.PrivateKey
	privKey, _ := btcec.PrivKeyFromBytes(privateKey) // Updated for btcec/v2
	wif, err := btcutil.NewWIF(privKey, network, true)
	if err != nil {
		log.Fatalf("❌ Failed to generate WIF: %v", err)
	}

	publicKey, err := btcutil.NewAddressPubKey(key.PublicKey().Key, network)
	if err != nil {
		log.Fatalf("❌ Failed to generate public key: %v", err)
	}

	fmt.Println("✅ New account created:")
	fmt.Println("🔑 Private Key:", hex.EncodeToString(privateKey))
	fmt.Println("🔑 WIF:", wif.String())
	fmt.Println("🏦 Address:", publicKey.EncodeAddress())
	fmt.Println("📝 Mnemonic:", mnemonic)

	return BitcoinAccount{
		PrivateKey: hex.EncodeToString(privateKey),
		WIF:        wif.String(),
		Address:    publicKey.EncodeAddress(),
	}
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func loadBitcoinAccount(wif string, isMainnet bool) BitcoinAccount {
	network := &chaincfg.MainNetParams
	if !isMainnet {
		network = &chaincfg.TestNet3Params
	}

	key, err := btcutil.DecodeWIF(wif)
	if err != nil {
		log.Fatalf("❌ Invalid WIF: %v", err)
	}

	publicKey, err := btcutil.NewAddressPubKey(key.PrivKey.PubKey().SerializeCompressed(), network)
	if err != nil {
		log.Fatalf("❌ Failed to generate public key: %v", err)
	}

	return BitcoinAccount{
		PrivateKey: hex.EncodeToString(key.PrivKey.Serialize()),
		WIF:        wif,
		Address:    publicKey.EncodeAddress(),
	}
}

// -------------------------------
// 💰 Get Account Balance
// -------------------------------
func getBitcoinBalance(apiURL, address string) float64 {
	resp, err := http.Get(fmt.Sprintf("%s/address/%s/utxo", apiURL, address))
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

	var utxos []BitcoinUTXOResponse
	if err := json.Unmarshal(body, &utxos); err != nil {
		log.Printf("❌ Failed to parse balance response: %v", err)
		return 0
	}

	var total int64
	for _, utxo := range utxos {
		total += utxo.Value
	}

	btcValue := float64(total) / 1e8 // Convert satoshis to BTC
	return btcValue
}

// -------------------------------
// 🚀 Send Transaction
// -------------------------------
func sendBitcoinTransaction(apiURL, wif, toAddress string, amountBTC float64, isMainnet bool) {
	network := &chaincfg.MainNetParams
	if !isMainnet {
		network = &chaincfg.TestNet3Params
	}

	key, err := btcutil.DecodeWIF(wif)
	if err != nil {
		log.Fatalf("❌ Invalid WIF: %v", err)
	}

	fromAddress, err := btcutil.NewAddressPubKey(key.PrivKey.PubKey().SerializeCompressed(), network)
	if err != nil {
		log.Fatalf("❌ Failed to generate from address: %v", err)
	}

	// Get UTXOs
	resp, err := http.Get(fmt.Sprintf("%s/address/%s/utxo", apiURL, fromAddress.EncodeAddress()))
	if err != nil {
		log.Fatalf("❌ Failed to get UTXOs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read UTXO response: %v", err)
	}

	var utxos []BitcoinUTXOResponse
	if err := json.Unmarshal(body, &utxos); err != nil {
		log.Fatalf("❌ Failed to parse UTXO response: %v", err)
	}

	if len(utxos) == 0 {
		log.Fatalf("❌ No UTXOs found for address %s", fromAddress.EncodeAddress())
	}

	// Create transaction
	tx := wire.NewMsgTx(wire.TxVersion)
	var totalInput int64
	for _, utxo := range utxos {
		hash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			log.Fatalf("❌ Invalid UTXO txid: %v", err)
		}
		txIn := wire.NewTxIn(&wire.OutPoint{Hash: *hash, Index: utxo.Vout}, nil, nil)
		tx.AddTxIn(txIn)
		totalInput += utxo.Value
	}

	// Add output
	amountSat := int64(amountBTC * 1e8)
	toAddr, err := btcutil.DecodeAddress(toAddress, network)
	if err != nil {
		log.Fatalf("❌ Invalid recipient address: %v", err)
	}
	toScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		log.Fatalf("❌ Failed to create output script: %v", err)
	}
	tx.AddTxOut(wire.NewTxOut(amountSat, toScript))

	// Add change output
	fee := int64(150 * 10) // Simplified: 10 sat/byte, 150 bytes
	change := totalInput - amountSat - fee
	if change > 0 {
		changeScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			log.Fatalf("❌ Failed to create change script: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(change, changeScript))
	}

	// Sign transaction
	for i, txIn := range tx.TxIn {
		// Generate the pkScript for the UTXO being spent (P2PKH script)
		pkScript, err := txscript.PayToAddrScript(fromAddress)
		if err != nil {
			log.Fatalf("❌ Failed to generate pkScript: %v", err)
		}
		sigScript, err := txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, key.PrivKey, true)
		if err != nil {
			log.Fatalf("❌ Failed to sign transaction: %v", err)
		}
		txIn.SignatureScript = sigScript
	}

	// Broadcast transaction
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		log.Fatalf("❌ Failed to serialize transaction: %v", err)
	}
	txHex := hex.EncodeToString(buf.Bytes())
	resp, err = http.Post(fmt.Sprintf("%s/tx", apiURL), "text/plain", bytes.NewBufferString(txHex))
	if err != nil {
		log.Fatalf("❌ Failed to broadcast transaction: %v", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read broadcast response: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 TxID: %s\n", string(body))
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// Networks
	networks := map[string]bool{
		"Bitcoin Mainnet": true,
		"Bitcoin Testnet": false,
	}

	// 1️⃣ Create a new account (or load existing)
	account := createBitcoinAccount(false) // Testnet
	// account := loadBitcoinAccount("YOUR_WIF", false)
	fmt.Println("\n🏦 Wallet Address:", account.Address)
	fmt.Println("\n🔑 Private Key:", account.PrivateKey)
	fmt.Println("\n🔑 WIF:", account.WIF)

	// 2️⃣ Check balances
	fmt.Println("\n💰 Balances:")
	for name, isMainnet := range networks {
		apiURL := connectBitcoinAPI(isMainnet)
		balance := getBitcoinBalance(apiURL, account.Address)
		fmt.Printf("%s: %.8f BTC\n", name, balance)
	}

	// 3️⃣ Example: Send 0.001 BTC (uncomment to test)
	// toAddress := "tb1..." // Replace with recipient address
	// sendBitcoinTransaction(connectBitcoinAPI(false), account.WIF, toAddress, 0.001, false)
}