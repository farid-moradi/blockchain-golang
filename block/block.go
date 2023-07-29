package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"goblockchain/transaction"
	"time"
)

type Block struct {
	nonce        int
	timestamp    int64
	previousHash [32]byte
	transactions []*transaction.Transaction
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Transactions() []*transaction.Transaction {
	return b.transactions
}

func SetBlockValues(nonce int, timestamp int64, previousHash [32]byte, transactions []*transaction.Transaction) *Block {
	b := new(Block)
	b.nonce = nonce
	b.previousHash = previousHash
	b.timestamp = timestamp
	b.transactions = transactions
	return b
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*transaction.Transaction) *Block {
	b := new(Block)
	b.nonce = nonce
	b.previousHash = previousHash
	b.timestamp = time.Now().UnixNano()
	b.transactions = transactions
	return b
}

func (b *Block) SetNonce(nonce int) {
	b.nonce = nonce
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce        int                        `json:"nonce"`
		Timestamp    int64                      `json:"timestamp"`
		PreviousHash string                     `json:"previous_hash"`
		Transactions []*transaction.Transaction `json:"transactions"`
	}{
		Nonce:        b.nonce,
		Timestamp:    b.timestamp,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	v := &struct {
		Timestamp    *int64                      `json:"timestamp"`
		Nonce        *int                        `json:"nonce"`
		PreviousHash *string                     `json:"previous_hash"`
		Transactions *[]*transaction.Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	return nil
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (b *Block) Print() {
	fmt.Printf("timestamp:          %d\n", b.timestamp)
	fmt.Printf("nonce:              %d\n", b.nonce)
	fmt.Printf("previousHash:       %x\n", b.previousHash)
	for _, t := range b.transactions {
		t.Print()
	}
}
