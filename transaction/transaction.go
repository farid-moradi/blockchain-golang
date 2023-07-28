package transaction

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func (t *Transaction) GetSenderBlockchainAddress() string {
	return t.senderBlockchainAddress
}

func (t *Transaction) GetRecipientBlockchainAddress() string {
	return t.recipientBlockchainAddress
}

func (t *Transaction) GetValueBlockchainAddress() float32 {
	return t.value
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address:     %s\n", t.senderBlockchainAddress)
	fmt.Printf(" recipient_blockchain_address:  %s\n", t.recipientBlockchainAddress)
	fmt.Printf(" value:                         %.1f\n", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	v := &struct {
		Sender    *string  `json:"sender_blockchain_address"`
		Recipient *string  `json:"recipient_blockchain_address"`
		Value     *float32 `json:"value"`
	}{
		Sender:    &t.senderBlockchainAddress,
		Recipient: &t.recipientBlockchainAddress,
		Value:     &t.value,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}
