package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// -------------------------------
// 🧬 Create a New Solana Account
// -------------------------------
func CreateAccount() (*solana.Wallet, solana.PublicKey) {
	wallet := solana.NewWallet()
	pubKey := wallet.PublicKey()

	fmt.Println("✅ New Solana account created:")
	fmt.Println("🔑 Private Key:", hex.EncodeToString(wallet.PrivateKey))
	fmt.Println("🏦 Address:", pubKey.String())
	return wallet, pubKey
}

// -------------------------------
// 🔐 Load Existing Account
// -------------------------------
func LoadAccount(privateKeyHex string) (*solana.PrivateKey, solana.PublicKey) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("❌ Invalid private key: %v", err)
	}
	privKey := solana.PrivateKey(privBytes)
	return &privKey, privKey.PublicKey()
}

// -------------------------------
// 💰 Get Solana Account Balance
// -------------------------------
func GetBalance(client *rpc.Client, publicKey solana.PublicKey) uint64 {
	balance, err := client.GetBalance(context.Background(), publicKey, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("❌ Failed to get balance: %v", err)
	}
	return balance.Value
}

// -------------------------------
// 🚀 Send SOL Transaction
// -------------------------------
func SendTransaction(client *rpc.Client, from *solana.PrivateKey, to solana.PublicKey, amountSOL float64) {
	amount := uint64(amountSOL * 1e9) // convert SOL to lamports

	recent, err := client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("❌ Failed to get recent blockhash: %v", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(amount, from.PublicKey(), to).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(from.PublicKey()),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create transaction: %v", err)
	}

	// Sign transaction
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key.Equals(from.PublicKey()) {
				return from
			}
			return nil
		},
	)
	if err != nil {
		log.Fatalf("❌ Failed to sign transaction: %v", err)
	}

	// Send transaction
	sig, err := client.SendTransaction(context.Background(), tx)
	if err != nil {
		log.Fatalf("❌ Failed to send transaction: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Signature: %s\n", sig.String())
}

// -------------------------------
// ⚙️ Utility Conversions
// -------------------------------
func LamportsToSOL(lamports uint64) float64 {
	return float64(lamports) / 1e9
}

func SOLToLamports(sol float64) uint64 {
	return uint64(sol * 1e9)
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// 🌐 Solana RPC Endpoints
	mainnets := map[string]string{
		"Solana Mainnet Beta": "https://api.mainnet-beta.solana.com",
	}

	testnets := map[string]string{
		"Solana Testnet": "https://api.testnet.solana.com",
		"Solana Devnet":  "https://api.devnet.solana.com",
	}

	// 1️⃣ Create a new account (or load existing)
	wallet, publicKey := CreateAccount()
	// To load an existing account:
	// walletPrivKey, publicKey := LoadAccount("YOUR_PRIVATE_KEY_HEX")

	fmt.Println("\n🏦 Wallet Address:", publicKey.String())
	fmt.Println("🔑 Private Key:", hex.EncodeToString(wallet.PrivateKey))

	// 2️⃣ Check balances on Mainnets
	fmt.Println("\n💰 Solana Mainnet Balances:")
	for name, rpcURL := range mainnets {
		client := rpc.New(rpcURL)
		balance := GetBalance(client, publicKey)
		fmt.Printf("%s: %f SOL\n", name, LamportsToSOL(balance))
	}

	// 3️⃣ Check balances on Testnets
	fmt.Println("\n💰 Solana Testnet Balances:")
	for name, rpcURL := range testnets {
		client := rpc.New(rpcURL)
		balance := GetBalance(client, publicKey)
		fmt.Printf("%s: %f SOL\n", name, LamportsToSOL(balance))
	}

	// 4️⃣ Example: Send 0.01 SOL to another address (Devnet)
	// ⚠️ Make sure the account has enough SOL on the target network
	recipient := "9phYaPEeniGzKRxLK2LkD5QgytBXhXQnrWmATLLGkvwN"

	toAddress, err := solana.PublicKeyFromBase58(recipient)
	if err != nil {
		log.Fatalf("❌ Invalid recipient address: %v", err)
	}

	SendTransaction(rpc.New("https://api.devnet.solana.com"), &wallet.PrivateKey, toAddress, 0.01)
}
