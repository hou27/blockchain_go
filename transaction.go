package main

const subsidy = 10

// Coin transaction
type Transaction struct {
	Txin  []TXInput
	Txout []TXOutput
}

// Transaction input
type TXInput struct {
	Txid      int
	Txout     int
	ScriptSig string
}

// Transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// Creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	txin := TXInput{-1, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{[]TXInput{txin}, []TXOutput{txout}}

	return &tx
}