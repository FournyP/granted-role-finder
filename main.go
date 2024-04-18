package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nanmu42/etherscan-api"
)

const (
	ROLE_GRANTED_TOPIC = "0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d"
	ROLE_REVOKED_TOPIC = "0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b"
)

func main() {
	var etherscanApiKey, etherscanBaseUrl, smartContractAddress string
	var fromBlock, toBlock int64

	flag.StringVar(&etherscanApiKey, "etherscan-api-key", "", "Etherscan API key.")
	flag.StringVar(&etherscanBaseUrl, "etherscan-base-url", "", "Etherscan base URL.")
	flag.StringVar(&smartContractAddress, "sc-address", "", "Smart contract address to scan.")

	flag.Int64Var(&toBlock, "to-block", 0, "To block number.")
	flag.Int64Var(&fromBlock, "from-block", 0, "From block number.")

	flag.Parse()

	if len(strings.TrimSpace(etherscanApiKey)) == 0 {
		panic("Etherscan API key is required.")
	}

	if len(strings.TrimSpace(etherscanBaseUrl)) == 0 {
		panic("Etherscan base URL is required.")
	}

	if len(strings.TrimSpace(smartContractAddress)) == 0 {
		panic("Smart contract address is required.")
	}

	if fromBlock == 0 {
		panic("From block number is required.")
	}

	if toBlock == 0 {
		panic("To block number is required.")
	}

	etherscanClient := etherscan.NewCustomized(etherscan.Customization{
		Timeout: 15 * time.Second,
		Key:     etherscanApiKey,
		BaseURL: etherscanBaseUrl,
		Verbose: false,
	})

	roleGrantedEthLogs, err := etherscanClient.GetLogs(int(fromBlock), int(toBlock), smartContractAddress, ROLE_GRANTED_TOPIC)
	if err != nil {
		log.Fatal("Unable to find role granted logs :", err)
	}

	log.Println("Role granted logs:")
	log.Println("-----------------------")

	for _, ethLog := range roleGrantedEthLogs {
		printLog(ethLog)
	}

	roleRevokedEthLogs, err := etherscanClient.GetLogs(int(fromBlock), int(toBlock), smartContractAddress, ROLE_REVOKED_TOPIC)
	if err != nil {
		log.Fatal("Unable to find role revoked logs :", err)
	}

	log.Println("Role revoked logs:")
	log.Println("-----------------------")

	for _, ethLog := range roleRevokedEthLogs {
		printLog(ethLog)
	}
}

func printLog(ethLog etherscan.Log) {
	role := ethLog.Topics[1]
	account := fmt.Sprintf("0x%s", ethLog.Topics[2][26:])
	sender := fmt.Sprintf("0x%s", ethLog.Topics[3][26:])

	log.Println("Role:", role)
	log.Println("Account:", account)
	log.Println("Sender:", sender)
	log.Println("-----------------------")
}
