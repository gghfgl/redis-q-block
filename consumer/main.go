package main

import (
	"fmt"
	"time"
	"log"
	"strconv"
	"crypto/sha256"
	"encoding/hex"

	"github.com/go-redis/redis/v7"
	"github.com/davecgh/go-spew/spew"
)

// NOTE: Block
type Block struct {
	Index     int // position of the data record in the blockchain
	Timestamp string
	Data       int    // pulse rate
	Hash      string // SHA256 id representing data record
	PrevHash  string // SHA256 id of the previous data record in the chain
}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.Data) + block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)

	return hex.EncodeToString(hashed)
}

func generateBlock(prevBlock Block, Data int) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = prevBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Data = Data
	newBlock.PrevHash = prevBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

func isBlockValid(block, prevBlock Block) bool {
	if prevBlock.Index+1 != block.Index {
		return false
	}

	if prevBlock.Hash != block.PrevHash {
		return false
	}

	if calculateHash(block) != block.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block, BlockChain []Block) {
	if len(newBlocks) > len(BlockChain) {
		BlockChain = newBlocks
	}
}


// NOTE: Main
const qKey = "blockQueue"

func main() {
	// Redis
	rdc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := rdc.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ping redis ok!")

	var BlockChain []Block
	t := time.Now()
	genesisBlock := Block{0, t.String(), 0, "", ""}
	spew.Dump(genesisBlock)
	BlockChain = append(BlockChain, genesisBlock)

	// consumer
	fmt.Println("Waiting for jobs on jobQueue:", qKey)
	for {
		time.Sleep(3 * time.Second)
		result, err := rdc.BLPop(0 * time.Second, qKey).Result()
		if err != nil {
			log.Fatal(err)
		}

		conv, err := strconv.Atoi(result[1])
		if err != nil {
			log.Printf("%v not a number: %v", result[1], err)
			continue
		}

		newBlock, err := generateBlock(BlockChain[len(BlockChain)-1], conv)
		if err != nil {
			log.Fatal(err)
		}

		if isBlockValid(newBlock, BlockChain[len(BlockChain)-1]) {
			BlockChain = append(BlockChain, newBlock)
			//replaceChain(newBlockchain, BlockChain)
			spew.Dump(newBlock)
		}

		fmt.Println("Executing job:", result[1])
	}
}
