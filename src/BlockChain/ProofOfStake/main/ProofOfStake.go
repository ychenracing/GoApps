package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
}

// 包含了所有通过验证的区块，即在整个区块链中已稳定的区块列表
var Blockchain []Block

// 区块副本
var tempBlocks []Block

// 包含了所有需要验证的区块。
var candidateBlocks = make(chan Block)

// 存储成功生成了区块的节点地址，然后把该地址广播给其他所有节点
var announcements = make(chan string)

var mutex = &sync.Mutex{}

// 每个节点抵押的token数量，key是节点地址，value是抵押的token数量
var validators = make(map[string]int)

func calculateHash(s string) string {
	sha := sha256.New()
	sha.Write([]byte(s))
	hashBytes := sha.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	return calculateHash(record)
}

func generateBlock(oldBlock Block, BPM int, address string) (Block, error) {
	result := &Block{
		Index:     oldBlock.Index + 1,
		Timestamp: time.Now().String(),
		BPM:       BPM,
		PrevHash:  oldBlock.Hash,
	}
	result.Hash = calculateBlockHash(*result)
	result.Validator = address
	return *result, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()
	// validator address
	var address string

	// allow user to allocate number of tokens to stake
	// the greater the number of tokens, the greater chance to forging a new block
	io.WriteString(conn, "Enter token balance:")
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now().String() + string(rand.Int31())

		// 服务器为每个客户端连接生成一个地址
		address = calculateHash(t)

		io.WriteString(conn, "\nYour address for stake: "+address)

		validators[address] = balance
		fmt.Println(validators)
		break
	}

	io.WriteString(conn, "\nEnter a new BPM:")

	scanBPM := bufio.NewScanner(conn)

	go func() {
		for {
			// take in BPM from stdin and add it to blockchain after conducting necessary validation
			for scanBPM.Scan() {
				bpm, err := strconv.Atoi(scanBPM.Text())
				// if malicious party tries to mutate the chain with a bad input, delete them as a validator and they lose their staked tokens
				if err != nil {
					log.Printf("%v not a number: %v", scanBPM.Text(), err)
					delete(validators, address)
					conn.Close()
				}

				mutex.Lock()
				oldLastIndex := Blockchain[len(Blockchain)-1]
				mutex.Unlock()

				// create newBlock for consideration to be forged
				newBlock, err := generateBlock(oldLastIndex, bpm, address)
				if err != nil {
					log.Println(err)
					continue
				}
				if isBlockValid(newBlock, oldLastIndex) {
					candidateBlocks <- newBlock
				}
				io.WriteString(conn, "\nEnter a new BPM:")
			}
		}
	}()

	// simulate receiving broadcast
	for {
		time.Sleep(time.Minute)
		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
	}
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if newBlock.Index != oldBlock.Index+1 {
		return false
	}
	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}
	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

// 服务器根据所有地址所质押的token数量，从所有候选区块中选择一个，加入到区块链尾部，然后像所有地址广播区块列表
func pickWinner2() {

	// 根据validators中质押的token数量，按照数量比率的概率选取其中一个
	mutex.Lock()
	tempValidators := validators
	tempBlocksCopy := tempBlocks // 质押了token并且通过验证了的block放入tempBlocks中
	mutex.Unlock()

	if len(tempBlocksCopy) <= 0 {
		return
	}

	var tokenCount int
	for _, v := range tempValidators {
		tokenCount += v
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	winnerIndex := r1.Intn(tokenCount)
	var stackCount int
	var winnerAddr string
	var winnerBlock Block
	for addr, v := range tempValidators {
		if winnerIndex >= stackCount && winnerIndex <= stackCount+v {
			winnerAddr = addr
			break
		}
		stackCount += v
	}

	for _, candBlock := range tempBlocksCopy {
		if candBlock.Validator == winnerAddr {
			winnerBlock = candBlock
			break
		}
	}

	mutex.Lock()
	Blockchain = append(Blockchain, winnerBlock)
	mutex.Unlock()
	for _ = range validators {
		// TODO: announcements是干什么用的
		announcements <- "\nwinning validator: " + winnerAddr + "\n"
	}

	mutex.Lock()
	// TODO: 从tempBlocks中选取了一个Block添加到主链上之后，tempBlocks置为空了，那些没选上的Block列表都被丢弃了？
	tempBlocks = []Block{}
	mutex.Unlock()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// create genesis block
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateBlockHash(genesisBlock), "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	// start TCP and serve TCP server
	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner2()
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

// pickWinner creates a lottery pool of validators and chooses the validator who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
//func pickWinner() {
//	time.Sleep(30 * time.Second)
//	mutex.Lock()
//	temp := tempBlocks
//	mutex.Unlock()
//
//	lotteryPool := make([]string, 0)
//	if len(temp) > 0 {
//
//		// slightly modified traditional proof of stake algorithm
//		// from all validators who submitted a block, weight them by the number of staked tokens
//		// in traditional proof of stake, validators can participate without submitting a block to be forged
//	OUTER:
//		for _, block := range temp {
//			// if already in lottery pool, skip
//			for _, node := range lotteryPool {
//				if block.Validator == node {
//					continue OUTER
//				}
//			}
//
//			// lock list of validators to prevent data race
//			mutex.Lock()
//			setValidators := validators
//			mutex.Unlock()
//
//			k, ok := setValidators[block.Validator]
//			if ok {
//				for i := 0; i < k; i++ {
//					lotteryPool = append(lotteryPool, block.Validator)
//				}
//			}
//		}
//
//		// randomly pick winner from lottery pool
//		s := rand.NewSource(time.Now().Unix())
//		r := rand.New(s)
//		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]
//
//		// add block of winner to blockchain and let all the other nodes know
//		for _, block := range temp {
//			if block.Validator == lotteryWinner {
//				mutex.Lock()
//				Blockchain = append(Blockchain, block)
//				mutex.Unlock()
//				for _ = range validators {
//					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
//				}
//				break
//			}
//		}
//	}
//
//	mutex.Lock()
//	tempBlocks = []Block{}
//	mutex.Unlock()
//}
