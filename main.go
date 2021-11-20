package main

func main() {
	chain := GetBlockchain()
	defer chain.db.Close()

	cli := Cli{chain}
	cli.Active()
}