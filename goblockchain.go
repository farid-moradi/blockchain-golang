package main

import (
	"fmt"
	"goblockchain/blockchain"
	"goblockchain/wallet"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	walletM := wallet.NewWallet()
	walletA := wallet.NewWallet()
	walletB := wallet.NewWallet()

	fmt.Println(walletM.BlockchainAddress())
	fmt.Println(walletA.BlockchainAddress())
	fmt.Println(walletB.BlockchainAddress())

	// Wallet
	t := wallet.NewTransaction(walletA.PrivateKey(), walletA.PublicKey(),
		walletA.BlockchainAddress(), walletB.BlockchainAddress(), 1.1)

	// Blockchain
	bc := blockchain.NewBlockChain(walletM.BlockchainAddress(), uint16(5000))
	sg := t.GenerateSignature()
	isAdded := bc.AddTransaction(walletA.BlockchainAddress(),
		walletB.BlockchainAddress(), 1.1, walletA.PublicKey(), sg)
	fmt.Println("added?", isAdded)

	bc.Mining()
	bc.Print()

	fmt.Printf("A %.1f\n", bc.CalculateTotalAmount(walletA.BlockchainAddress()))
	fmt.Printf("B %.1f\n", bc.CalculateTotalAmount(walletB.BlockchainAddress()))
	fmt.Printf("M %.1f\n", bc.CalculateTotalAmount(walletM.BlockchainAddress()))

}
