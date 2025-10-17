package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// ConnectLitecoinClient connects to a Litecoin network
func ConnectLitecoinClient(rpcURL, user, pass string) *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		Host:         rpcURL,
		User:         user,
		Pass:         pass,
		HTTPPostMode: true,
		DisableTLS:   false, // Enable TLS for HTTPS
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Litecoin network: %v", err)
	}
	return client
}

// CreateLitecoinAccount generates a new Litecoin account
func CreateLitecoinAccount(net *chaincfg.Params) (wif *btcutil.WIF, address string) {
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		log.Fatalf("❌ Failed to generate private key: %v", err)
	}

	wifObj, err := btcutil.NewWIF(privKey, net, true)
	if err != nil {
		log.Fatalf("❌ Failed to create WIF: %v", err)
	}

	pubKeyHash := btcutil.Hash160(privKey.PubKey().SerializeCompressed())
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, net)
	if err != nil {
		log.Fatalf("❌ Failed to create address: %v", err)
	}

	address = addr.EncodeAddress()

	fmt.Println("✅ New Litecoin account created:")
	fmt.Println("🔑 WIF:", wifObj.String())
	fmt.Println("🏦 Address:", address)

	return wifObj, address
}

// LoadLitecoinAccount loads an existing Litecoin account from a WIF
func LoadLitecoinAccount(wifStr string, net *chaincfg.Params) (*btcutil.WIF, string, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, "", fmt.Errorf("❌ Invalid WIF: %v", err)
	}

	pubKeyHash := btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed())
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, net)
	if err != nil {
		return nil, "", fmt.Errorf("❌ Failed to create address: %v", err)
	}

	return wif, addr.EncodeAddress(), nil
}

// GetLitecoinBalance retrieves the balance of a Litecoin account
func GetLitecoinBalance(client *rpcclient.Client, address string) *big.Float {
	// Placeholder: returns 0 as accounts are new and public APIs like BlockCypher are used
	// In a real implementation, use HTTP requests to query balance via the API
	return big.NewFloat(0.0)
}

// SendLitecoinTransaction sends a Litecoin transaction
func SendLitecoinTransaction(client *rpcclient.Client, wif *btcutil.WIF, toAddress string, amountBTC float64, net *chaincfg.Params) {
	fromAddr, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(wif.PrivKey.PubKey().SerializeCompressed()), net)
	if err != nil {
		log.Fatalf("❌ Failed to create from address: %v", err)
	}

	toAddr, err := btcutil.DecodeAddress(toAddress, net)
	if err != nil {
		log.Fatalf("❌ Invalid to address: %v", err)
	}

	amount := btcutil.Amount(amountBTC * 1e8) // Convert to satoshis

	// Get unspent outputs
	utxos, err := client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{fromAddr})
	if err != nil {
		log.Fatalf("❌ Failed to list unspent: %v", err)
	}

	if len(utxos) == 0 {
		log.Fatal("❌ No unspent outputs available")
	}

	// Create transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	totalInput := btcutil.Amount(0)
	for _, utxo := range utxos {
		txid, err := hex.DecodeString(utxo.TxID)
		if err != nil {
			log.Fatalf("❌ Invalid txid: %v", err)
		}
		var hash chainhash.Hash
		copy(hash[:], txid)
		outPoint := wire.NewOutPoint(&hash, utxo.Vout)
		txIn := wire.NewTxIn(outPoint, nil, nil)
		tx.AddTxIn(txIn)
		totalInput += btcutil.Amount(utxo.Amount * 1e8)
		if totalInput >= amount+1000 { // + fee
			break
		}
	}

	if totalInput < amount+1000 {
		log.Fatal("❌ Insufficient funds")
	}

	// Add output
	pkScript, err := txscript.PayToAddrScript(toAddr)
	if err != nil {
		log.Fatalf("❌ Failed to create pkScript: %v", err)
	}
	tx.AddTxOut(wire.NewTxOut(int64(amount), pkScript))

	// Change output
	change := totalInput - amount - 1000
	if change > 0 {
		changeScript, err := txscript.PayToAddrScript(fromAddr)
		if err != nil {
			log.Fatalf("❌ Failed to create change script: %v", err)
		}
		tx.AddTxOut(wire.NewTxOut(int64(change), changeScript))
	}

	// Sign transaction
	for i, txIn := range tx.TxIn {
		scriptPubKey, err := hex.DecodeString(utxos[i].ScriptPubKey)
		if err != nil {
			log.Fatalf("❌ Failed to decode scriptPubKey: %v", err)
		}
		sigScript, err := txscript.SignatureScript(tx, i, scriptPubKey, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			log.Fatalf("❌ Failed to sign: %v", err)
		}
		txIn.SignatureScript = sigScript
	}

	// Send transaction
	txHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 TxID: %s\n", txHash.String())
}

// SatoshisToLTC converts Satoshis to LTC
func SatoshisToLTC(satoshis int64) *big.Float {
	return new(big.Float).Quo(big.NewFloat(float64(satoshis)), big.NewFloat(1e8))
}

// LTCToSatoshis converts LTC to Satoshis
func LTCToSatoshis(ltc float64) int64 {
	return int64(ltc * 1e8)
}

func main() {
	// Litecoin network configurations
	litecoinNetworks := map[string]map[string]string{
		"Litecoin Mainnet": {
			"rpc":  "https://api.blockcypher.com/v1/ltc/main", // Public API, no auth needed
			"user": "",
			"pass": "",
		},
		"Litecoin Testnet": {
			"rpc":  "https://api.blockcypher.com/v1/ltc/test3", // Public API, no auth needed
			"user": "",
			"pass": "",
		},
	}

	// Define Litecoin network parameters
	netParams := &chaincfg.Params{
		Name:             "litecoin",
		PubKeyHashAddrID: 0x30, // Litecoin mainnet P2PKH address prefix (L)
		ScriptHashAddrID: 0x32, // Litecoin mainnet P2SH address prefix (M)
		PrivateKeyID:     0xB0, // Litecoin private key prefix
	}

	// Create a new Litecoin account
	wif, address := CreateLitecoinAccount(netParams)
	fmt.Println("\n🏦 Litecoin Wallet Address:", address)
	fmt.Println("\n🏦 Litecoin WIF:", wif.String())

	// Check balances on Litecoin networks
	fmt.Println("\n💰 Litecoin Balances:")
	for name, config := range litecoinNetworks {
		client := ConnectLitecoinClient(config["rpc"], config["user"], config["pass"])
		balance := GetLitecoinBalance(client, address)
		fmt.Printf("%s: %f LTC\n", name, balance)
	}
}