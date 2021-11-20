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
	ScriptSig string // Release script
}

// Transaction output
type TXOutput struct {
	Value        int
	ScriptPubKey string // Lock script
}

// Creates a new coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	txin := TXInput{-1, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{[]TXInput{txin}, []TXOutput{txout}}

	return &tx
}

func NewUTXOTransaction(from, to string, amount int) *Transaction {

	return &Transaction{}
}