package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/FournyP/granted-role-finder/abis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nanmu42/etherscan-api"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

func main() {
	var rpcUrl, etherscanApiKey, etherscanBaseUrl, smartContractAddress, topic string
	var fromBlock, toBlock int64

	flag.StringVar(&rpcUrl, "rpc-url", "", "RPC URL of the Ethereum node.")
	flag.StringVar(&etherscanApiKey, "etherscan-api-key", "", "Etherscan API key.")
	flag.StringVar(&etherscanBaseUrl, "etherscan-base-url", "", "Etherscan base URL.")
	flag.StringVar(&smartContractAddress, "sc-address", "", "Smart contract address to scan.")
	flag.StringVar(&topic, "topic", "0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d", "Topic to filter logs.")

	flag.Int64Var(&toBlock, "to-block", 0, "To block number.")
	flag.Int64Var(&fromBlock, "from-block", 0, "From block number.")

	flag.Parse()

	if len(strings.TrimSpace(rpcUrl)) == 0 {
		panic("RPC URL is required.")
	}

	if len(strings.TrimSpace(etherscanApiKey)) == 0 {
		panic("Etherscan API key is required.")
	}

	if len(strings.TrimSpace(etherscanBaseUrl)) == 0 {
		panic("Etherscan base URL is required.")
	}

	if len(strings.TrimSpace(smartContractAddress)) == 0 {
		panic("Smart contract address is required.")
	}

	if len(strings.TrimSpace(topic)) == 0 {
		panic("Topic is required.")
	}

	ethClient, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	etherscanClient := etherscan.NewCustomized(etherscan.Customization{
		Timeout: 15 * time.Second,
		Key:     etherscanApiKey,
		BaseURL: etherscanBaseUrl,
		Verbose: false,
	})

	if toBlock == 0 {
		header, err := ethClient.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}

		toBlock = header.Number.Int64()
	}

	ethLogs, err := etherscanClient.GetLogs(int(fromBlock), int(toBlock), smartContractAddress, topic)
	if err != nil {
		log.Fatal("Unable to find logs :", err)
	}

	hashs := lop.Map(ethLogs, func(value etherscan.Log, _ int) string {
		return value.TransactionHash
	})

	hashs = lo.Uniq(hashs)

	for _, hash := range hashs {
		txHash := common.HexToHash(hash)
		receipt, err := ethClient.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			log.Fatal("Unable to get transaction receipt : ", hash, err)
		}

		for _, ethLog := range receipt.Logs {
			accessControl, err := abis.NewAccessControl(ethLog.Address, ethClient)
			if err != nil {
				log.Fatal("Unable to get AccessControl contract : ", err)
			}

			event, err := accessControl.ParseRoleGranted(*ethLog)
			if err != nil {
				log.Printf("Failed to unpack log: %v", err)
				continue
			}

			log.Printf("Role: %s", event)
		}
	}
}
