# 🌐 Multi-Chain Go Scripts

A collection of open-source **Golang scripts** for interacting with multiple blockchains.  
Supports **account creation**, **balance retrieval**, and **token/coin transfers** across major Layer 1 and Layer 2 networks — both mainnet and testnet.

Ideal for developers building Web3 applications, testing blockchain features, or learning multi-chain development with Go.

---

## 🚀 Features

- ✅ Account creation
- ✅ Balance check
- ✅ Native token transfers
- ✅ Modular & clean Go code
- ✅ Easy to integrate into your own projects

---

## 🔗 Supported Blockchains

| Blockchain       | Mainnet           | Testnet              |
|------------------|-------------------|----------------------|
| Ethereum         | ✅ Mainnet         | ✅ Sepolia            |
| Polygon          | ✅ Mainnet         | ✅ Amoy               |
| BNB Smart Chain  | ✅ Mainnet         | ✅ Testnet            |
| Base             | ✅ Mainnet         | ✅ Sepolia            |
| Celo             | ✅ Mainnet         | ✅ Alfajores          |
| Optimism         | ✅ Mainnet         | ✅ Sepolia            |
| Linea            | ✅ Mainnet         | ✅ Sepolia            |
| Avalanche        | ✅ Mainnet         | ✅ Fuji               |
| Solana           | ✅ Mainnet         | ✅ Devnet             |
| Tron             | ✅ Mainnet         | ✅ Shasta             |
| Bitcoin          | ✅ Mainnet         | ✅ Testnet            |
| Stacks           | ✅ Mainnet         | ✅ Testnet            |
| Eclipse          | ✅ Mainnet         | ✅ Testnet            |
| TON              | ✅ Mainnet         | ✅ Testnet            |
| **Aptos**        | ✅ Mainnet         | ✅ Testnet            |
| **Sui**          | ✅ Mainnet         | ✅ Testnet            |

---

## 🛠 Usage

Each blockchain has its own folder containing:
- `create_account.go` – creates a new wallet/account
- `get_balance.go` – fetches the balance of an address
- `transfer.go` – sends tokens/coins to another address

You can run any script using:

```bash
go run ./<blockchain>/<script>.go
```
