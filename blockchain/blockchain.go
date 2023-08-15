package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"goblockchain/block"
	"goblockchain/transaction"
	"goblockchain/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MINING_DIFFICULTY                = 3
	MINING_SENDER                    = "THE BLOCKCHAIN"
	MINING_REWARD                    = 1.0
	MINING_TIMER_SEC                 = 20
	BLOCKCHAIN_PORT_RANGE_START      = 5000
	BLOCKCHAIN_PORT_RANGE_END        = 5003
	NEIGHBOR_IP_RANGE_START          = 0
	NEIGHBOR_IP_RANGE_END            = 1
	BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC = 20
)

type Blockchain struct {
	transactionPool   []*transaction.Transaction
	chain             []*block.Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
	muxNeighbors      sync.Mutex
	neighbors         []string
}

func NewBlockChain(blockchainAddress string, port uint16) *Blockchain {
	b := &block.Block{}
	bc := new(Blockchain)
	bc.blockchainAddress = blockchainAddress
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

func (bc *Blockchain) Chain() []*block.Block {
	return bc.chain
}

func (bc *Blockchain) Run() {
	bc.StartSyncNeighbors()
	// immediatelly after starting a blockchain server, we need to find other neighbors
	// after that every server that has been started need to get the correct chain from
	// other blockchain servers by calling the resolve conflict method
	bc.ResolveConflicts()
	bc.StartMining()
}

func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(utils.GetHost(), bc.port, NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)

	log.Printf("%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) TransactionPool() []*transaction.Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *block.Block {
	b := block.NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*transaction.Transaction{}
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
	return b
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*block.Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Blocks *[]*block.Block `json:"chain"`
	}{
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) CopyTransactionPool() []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions, transaction.NewTransaction(t.GetSenderBlockchainAddress(), t.GetRecipientBlockchainAddress(), t.GetValueBlockchainAddress()))
	}
	return transactions
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce++
	}
	return nonce
}

func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	// if we don't use this if statement, a miner can get rewarder even when there is no transaction
	// we are going to comment this if statement off, because in real world mining can be
	// preform with 0 transaction and of course it is not likely to happen
	// if len(bc.transactionPool) == 0 {
	// 	return false
	// }
	bc.AddTransaction(MINING_SENDER, bc.blockchainAddress, MINING_REWARD, nil, nil)
	previousHash := bc.LastBlock().Hash()
	nonce := bc.ProofOfWork()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	return true
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

func (bc *Blockchain) CalculateTotalAmount(blockchainAddress string) float32 {
	var totalAmount float32
	for _, v := range bc.chain {
		for _, t := range v.Transactions() {
			if t.GetSenderBlockchainAddress() == blockchainAddress {
				totalAmount -= t.GetValueBlockchainAddress()
			} else if t.GetRecipientBlockchainAddress() == blockchainAddress {
				totalAmount += t.GetValueBlockchainAddress()
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) ValidChain(chain []*block.Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.PreviousHash() != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), MINING_DIFFICULTY) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) LastBlock() *block.Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*transaction.Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := block.NewBlockWithoutTime(nonce, previousHash, transactions)
	guessBlockHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessBlockHashStr[:difficulty] == zeros
}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)

	if isTransacted {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &recipient, &publicKeyStr, &value, &signatureStr}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("%v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) AddTransaction(sender string, recipient string, value float32,
	senderPublickey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := transaction.NewTransaction(sender, recipient, value)

	// for the mining node getting its reward from the network, there is no need to verify the reward transaction
	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublickey, s, t) {
		if bc.CalculateTotalAmount(sender) < value {
			log.Println("ERROR: Not enough balance in a wallet")
			return false
		}
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: Verify Transaction")
	}
	return false
}

func (bc *Blockchain) VerifyTransactionSignature(senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature, t *transaction.Transaction) bool {
	t.Print()
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	verified := ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
	fmt.Println(verified)
	return verified
}

func (bc *Blockchain) Print() {
	for i, v := range bc.chain {
		fmt.Printf("%sChain %d%s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		v.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string  `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*block.Block = nil
	maxLength := len(bc.chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, _ := http.Get(endpoint)
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			_ = decoder.Decode(&bcResp)

			chain := bcResp.Chain()

			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle confilicts replaced")
		return true
	}
	log.Printf("Resovle conflicts not replaced")
	return false
}
