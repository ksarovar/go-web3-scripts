package main

import (
	"fmt"
	"log"
	"math/big"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// ConnectSubstrateClient connects to a Polkadot/Substrate network
func ConnectSubstrateClient(rpcURL string) *gsrpc.SubstrateAPI {
	api, err := gsrpc.NewSubstrateAPI(rpcURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Polkadot network: %v", err)
	}
	return api
}

// CreatePolkadotAccount generates a new Polkadot account
func CreatePolkadotAccount() (mnemonic string, address string) {
	// Using a fixed mnemonic for demonstration; in practice, generate a new one
	mnemonic = "bottom drive obey lake curtain smoke basket hold race lonely fit walk"
	keyringPair, err := signature.KeyringPairFromSecret(mnemonic, 42)
	if err != nil {
		log.Fatalf("❌ Failed to generate from mnemonic: %v", err)
	}

	address = keyringPair.Address

	fmt.Println("✅ New Polkadot account created:")
	fmt.Println("🔑 Mnemonic:", mnemonic)
	fmt.Println("🏦 Address:", address)

	return mnemonic, address
}

// LoadPolkadotAccount loads an existing Polkadot account from a mnemonic
func LoadPolkadotAccount(mnemonic string) (signature.KeyringPair, error) {
	keyringPair, err := signature.KeyringPairFromSecret(mnemonic, 42)
	if err != nil {
		return signature.KeyringPair{}, fmt.Errorf("❌ Invalid mnemonic: %v", err)
	}
	return keyringPair, nil
}

// GetPolkadotBalance retrieves the balance of a Polkadot account (placeholder)
func GetPolkadotBalance(api *gsrpc.SubstrateAPI, address string) *big.Float {
	// Placeholder: returns 0 as the account is new and likely has no balance
	// In a real implementation, decode SS58 address and query the chain
	return big.NewFloat(0.0)
}

// SendPolkadotTransaction sends a Polkadot transaction
func SendPolkadotTransaction(api *gsrpc.SubstrateAPI, keyringPair signature.KeyringPair, toAddress string, amountDOT float64) {
	amountPlancks := uint64(amountDOT * 1e10)

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		log.Fatalf("❌ Failed to get metadata: %v", err)
	}

	toAddr, err := types.NewAddressFromHexAccountID(toAddress)
	if err != nil {
		log.Fatalf("❌ Invalid to address: %v", err)
	}
	call, err := types.NewCall(meta, "Balances.transfer", toAddr.AsAccountID, types.NewUCompactFromUInt(amountPlancks))
	if err != nil {
		log.Fatalf("❌ Failed to create call: %v", err)
	}

	extrinsic := types.NewExtrinsic(call)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		log.Fatalf("❌ Failed to get genesis hash: %v", err)
	}

	runtimeVersion, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		log.Fatalf("❌ Failed to get runtime version: %v", err)
	}

	fromAddr, err := types.NewAddressFromHexAccountID(keyringPair.Address)
	if err != nil {
		log.Fatalf("❌ Invalid from address: %v", err)
	}
	key, err := types.CreateStorageKey(meta, "System", "Account", fromAddr.AsAccountID.ToBytes(), nil)
	if err != nil {
		log.Fatalf("❌ Failed to create storage key: %v", err)
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		log.Fatalf("❌ Failed to get account info: %v", err)
	}

	nonce := uint32(accountInfo.Nonce)

	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        runtimeVersion.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		TransactionVersion: runtimeVersion.TransactionVersion,
	}

	err = extrinsic.Sign(keyringPair, o)
	if err != nil {
		log.Fatalf("❌ Failed to sign extrinsic: %v", err)
	}

	hash, err := api.RPC.Author.SubmitExtrinsic(extrinsic)
	if err != nil {
		log.Fatalf("❌ Failed to submit extrinsic: %v", err)
	}

	fmt.Printf("✅ Transaction sent successfully!\n🔗 Hash: %s\n", hash.Hex())
}

// PlancksToDOT converts Plancks to DOT
func PlancksToDOT(plancks uint64) *big.Float {
	return new(big.Float).Quo(big.NewFloat(float64(plancks)), big.NewFloat(1e10))
}

// DOTToPlancks converts DOT to Plancks
func DOTToPlancks(dot float64) uint64 {
	return uint64(dot * 1e10)
}

func main() {
	// Polkadot network configurations
	polkadotNetworks := map[string]string{
		"Polkadot Mainnet":         "wss://rpc.polkadot.io",
		"Polkadot Westend Testnet": "wss://westend-rpc.polkadot.io",
	}

	// Create a new Polkadot account
	mnemonic, address := CreatePolkadotAccount()
	fmt.Println("\n🏦 Polkadot Wallet Address:", address)
	fmt.Println("\n🏦 Polkadot Mnemonic Phrase:", mnemonic)

	// Check balances on Polkadot networks
	fmt.Println("\n💰 Polkadot Balances:")
	for name, rpc := range polkadotNetworks {
		api := ConnectSubstrateClient(rpc)
		balance := GetPolkadotBalance(api, address)
		fmt.Printf("%s: %f DOT\n", name, balance)
	}
}