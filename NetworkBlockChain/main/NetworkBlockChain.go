package main

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	"github.com/joho/godotenv"
	"log"
	"github.com/davecgh/go-spew/spew"
	"net"
	"os"
	"io"
	"bufio"
	"strconv"
	"encoding/json"
)

type NetworkBlock struct {
	Index int
	Timestamp string
	BPM int
	Hash string
	PrevHash string
}

var Blockchain []NetworkBlock

var bcServer chan []NetworkBlock

func calculateHash(newBlock NetworkBlock) string {
	record := string(newBlock.Index) + newBlock.Timestamp + string(newBlock.BPM) + newBlock.PrevHash
	sha := sha256.New()
	sha.Write([]byte(record))
	hashBytes := sha.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func generateBlock(oldBlock NetworkBlock, BPM int) (NetworkBlock, error) {
	var newBlock NetworkBlock
	t := time.Now()
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)
	return newBlock, nil
}

func isBlockValid(newBlock, oldBlock NetworkBlock) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

func replaceChain(newBlocks []NetworkBlock) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	io.WriteString(conn, "Enter a new BPM:")
	scanner := bufio.NewScanner(conn)

	// 每来一个新连接，服务端会创造2个无限循环的goroutine，服务端压力很大
	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}
			newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}

			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(conn, string(output))
		}
	}()
	for _ = range bcServer {
		spew.Dump(Blockchain)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []NetworkBlock)

	t := time.Now()
	genesisBlock := NetworkBlock{0, t.String(), 0, "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	server, err := net.Listen("tcp", ":"+os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}