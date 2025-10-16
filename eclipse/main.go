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
// 🔗 Connect to RPC
// -------------------------------
func ConnectClient(rpcURL string) *rpc.Client {
	client := rpc.New(rpcURL)
	_, err := client.GetVersion(context.Background())
	if err != nil {
		log.Fatalf("❌ Failed to connect to Eclipse network: %v", err)
	}
	return client
}

// -------------------------------
// 🧬 Create a New Eclipse Account
// -------------------------------
func CreateAccount() (*solana.Wallet, solana.PublicKey) {
	wallet := solana.NewWallet()
	pubKey := wallet.PublicKey()

	fmt.Println("✅ New Eclipse account created:")
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
// 💰 Get Eclipse Account Balance
// -------------------------------
func GetBalance(client *rpc.Client, publicKey solana.PublicKey) uint64 {
	balance, err := client.GetBalance(context.Background(), publicKey, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("❌ Failed to get balance: %v", err)
	}
	return balance.Value
}

// -------------------------------
// 🚀 Send ECL Transaction
// -------------------------------
func SendTransaction(client *rpc.Client, from *solana.PrivateKey, to solana.PublicKey, amountECL float64) {
	amount := uint64(amountECL * 1e9) // Convert ECL to lamports (assuming 1 ECL = 10^9 lamports, similar to SOL)

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
func LamportsToECL(lamports uint64) float64 {
	return float64(lamports) / 1e9 // Assuming 1 ECL = 10^9 lamports
}

func ECLToLamports(ecl float64) uint64 {
	return uint64(ecl * 1e9)
}

// -------------------------------
// 🧩 Main Example with Multi-network Balance Check
// -------------------------------
func main() {
	// -------------------------------
	// 🌐 Eclipse RPC Endpoints
	// -------------------------------
	// Note: Eclipse RPC endpoints may not be publicly available yet or may require specific access.
	// Replace with actual Eclipse mainnet/testnet RPC URLs when available.
	// For now, placeholders are used based on typical Solana-compatible RPC structure.
	mainnets := map[string]string{
		"Eclipse Mainnet": "https://mainnetbeta-rpc.eclipse.xyz", // Placeholder, replace with actual Eclipse Mainnet RPC
	}

	testnets := map[string]string{
		"Eclipse Testnet": "https://testnet.dev2.eclipsenetwork.xyz", // Placeholder, replace with actual Eclipse Testnet RPC
		"Eclipse Devnet":  "https://staging-rpc.dev2.eclipsenetwork.xyz",  // Placeholder, replace with actual Eclipse Devnet RPC
	}

	// 1️⃣ Create a new account (or load existing)
	wallet, publicKey := CreateAccount()
	// To load an existing account:
	// walletPrivKey, publicKey := LoadAccount("YOUR_PRIVATE_KEY_HEX")

	fmt.Println("\n🏦 Wallet Address:", publicKey.String())
	fmt.Println("🔑 Private Key:", hex.EncodeToString(wallet.PrivateKey))

	// 2️⃣ Check balances on Mainnets
	fmt.Println("\n💰 Eclipse Mainnet Balances:")
	for name, rpcURL := range mainnets {
		client := ConnectClient(rpcURL)
		balance := GetBalance(client, publicKey)
		fmt.Printf("%s: %f ECL\n", name, LamportsToECL(balance))
	}

	// 3️⃣ Check balances on Testnets
	fmt.Println("\n💰 Eclipse Testnet Balances:")
	for name, rpcURL := range testnets {
		client := ConnectClient(rpcURL)
		balance := GetBalance(client, publicKey)
		fmt.Printf("%s: %f ECL\n", name, LamportsToECL(balance))
	}

	// 4️⃣ Example: Send 0.01 ECL to another address (Testnet)
	// ⚠️ Make sure the account has enough ECL on the target network
	recipient := "9phYaPEeniGzKRxLK2LkD5QgytBXhXQnrWmATLLGkvwN" // Replace with a valid Eclipse address

	toAddress, err := solana.PublicKeyFromBase58(recipient)
	if err != nil {
		log.Fatalf("❌ Invalid recipient address: %v", err)
	}

	client := ConnectClient("https://testnet.dev2.eclipsenetwork.xyz") // Replace with actual Testnet RPC
	SendTransaction(client, &wallet.PrivateKey, toAddress, 0.01)
}