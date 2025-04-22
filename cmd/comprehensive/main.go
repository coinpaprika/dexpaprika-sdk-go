package main

import (
	"context"
	"fmt"
	"time"

	"github.com/coinpaprika/dexpaprika-sdk-go/dexpaprika"
)

func main() {
	// Create a new client with default options
	client := dexpaprika.NewClient()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// For storing test results
	results := make(map[string]bool)
	var testName string

	fmt.Println("🧪 RUNNING COMPREHENSIVE TESTS OF DEXPAPRIKA SDK")
	fmt.Println("===============================================")

	// -------------------------
	// Test NetworksService
	// -------------------------
	fmt.Println("\n📝 TESTING NetworksService")

	// 1. List networks
	testName = "NetworksService.List"
	fmt.Printf("Testing %s... ", testName)
	networks, err := client.Networks.List(ctx)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Found %d networks\n", len(networks))
		results[testName] = true

		// Store a network ID for later tests if available
		var networkID string
		if len(networks) > 0 {
			networkID = networks[0].ID
		} else {
			networkID = "ethereum" // Fallback to ethereum if no networks returned
		}

		// 2. List DEXes on a network
		testName = "NetworksService.ListDexes"
		fmt.Printf("Testing %s (network=%s)... ", testName, networkID)
		dexes, err := client.Networks.ListDexes(ctx, networkID, 0, 5)
		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			results[testName] = false
		} else {
			fmt.Printf("✅ SUCCESS: Found %d DEXes\n", len(dexes.Dexes))
			results[testName] = true
		}
	}

	// -------------------------
	// Test PoolsService
	// -------------------------
	fmt.Println("\n📝 TESTING PoolsService")

	// 1. List pools
	testName = "PoolsService.List"
	fmt.Printf("Testing %s... ", testName)
	poolsOpts := &dexpaprika.ListOptions{
		Limit: 5,
	}
	pools, err := client.Pools.List(ctx, poolsOpts)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Found %d pools\n", len(pools.Pools))
		results[testName] = true

		// Store pool info for later tests if available
		var poolID, poolChain, dexID string
		if len(pools.Pools) > 0 {
			poolID = pools.Pools[0].ID
			poolChain = pools.Pools[0].Chain
			dexID = pools.Pools[0].DexID

			// 2. List pools by network
			testName = "PoolsService.ListByNetwork"
			fmt.Printf("Testing %s (network=%s)... ", testName, poolChain)
			networkPools, err := client.Pools.ListByNetwork(ctx, poolChain, poolsOpts)
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				results[testName] = false
			} else {
				fmt.Printf("✅ SUCCESS: Found %d pools\n", len(networkPools.Pools))
				results[testName] = true
			}

			// 3. List pools by DEX
			testName = "PoolsService.ListByDex"
			fmt.Printf("Testing %s (network=%s, dex=%s)... ", testName, poolChain, dexID)
			dexPools, err := client.Pools.ListByDex(ctx, poolChain, dexID, poolsOpts)
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				results[testName] = false
			} else {
				fmt.Printf("✅ SUCCESS: Found %d pools\n", len(dexPools.Pools))
				results[testName] = true
			}

			// 4. Get pool details
			testName = "PoolsService.GetDetails"
			fmt.Printf("Testing %s (network=%s, pool=%s)... ", testName, poolChain, poolID)
			poolDetails, err := client.Pools.GetDetails(ctx, poolChain, poolID, false)
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				results[testName] = false
			} else {
				fmt.Printf("✅ SUCCESS: Got pool details (Fee: %.2f%%)\n", poolDetails.Fee*100)
				results[testName] = true
			}

			// 5. Get pool OHLCV data
			testName = "PoolsService.GetOHLCV"
			fmt.Printf("Testing %s (network=%s, pool=%s)... ", testName, poolChain, poolID)
			now := time.Now()
			yesterday := now.Add(-24 * time.Hour)
			ohlcvOpts := &dexpaprika.OHLCVOptions{
				Start:    yesterday.Format("2006-01-02"),
				End:      now.Format("2006-01-02"),
				Interval: "1h",
				Limit:    24,
			}
			ohlcv, err := client.Pools.GetOHLCV(ctx, poolChain, poolID, ohlcvOpts)
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				results[testName] = false
			} else {
				fmt.Printf("✅ SUCCESS: Got %d OHLCV records\n", len(ohlcv))
				results[testName] = true
			}

			// 6. Get pool transactions
			testName = "PoolsService.GetTransactions"
			fmt.Printf("Testing %s (network=%s, pool=%s)... ", testName, poolChain, poolID)
			transactions, err := client.Pools.GetTransactions(ctx, poolChain, poolID, 0, 5, "")
			if err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				results[testName] = false
			} else {
				fmt.Printf("✅ SUCCESS: Got %d transactions\n", len(transactions.Transactions))
				results[testName] = true
			}
		} else {
			fmt.Println("⚠️ Skipping other PoolsService tests - no pools found")
		}
	}

	// -------------------------
	// Test TokensService
	// -------------------------
	fmt.Println("\n📝 TESTING TokensService")

	// For token tests, we'll try to use BTC if we can find it in search
	testName = "SearchService.Search for token tests"
	fmt.Printf("Finding token for token tests... ")
	searchResults, err := client.Search.Search(ctx, "bitcoin")
	var tokenChain, tokenAddress string
	if err != nil || len(searchResults.Tokens) == 0 {
		fmt.Printf("⚠️ WARNING: Could not find token for tests, using fallback values\n")
		// Fallback to Ethereum WETH address
		tokenChain = "ethereum"
		tokenAddress = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	} else {
		fmt.Printf("Found token: %s (%s on %s)\n",
			searchResults.Tokens[0].Name,
			searchResults.Tokens[0].Symbol,
			searchResults.Tokens[0].Chain)
		tokenChain = searchResults.Tokens[0].Chain
		tokenAddress = searchResults.Tokens[0].ID
	}

	// 1. Get token details
	testName = "TokensService.GetDetails"
	fmt.Printf("Testing %s (network=%s, token=%s)... ", testName, tokenChain, tokenAddress)
	tokenDetails, err := client.Tokens.GetDetails(ctx, tokenChain, tokenAddress)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Got details for %s (%s)\n", tokenDetails.Name, tokenDetails.Symbol)
		results[testName] = true
	}

	// 2. Get token pools
	testName = "TokensService.GetPools"
	fmt.Printf("Testing %s (network=%s, token=%s)... ", testName, tokenChain, tokenAddress)
	tokenPools, err := client.Tokens.GetPools(ctx, tokenChain, tokenAddress, poolsOpts, "")
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Found %d pools for token\n", len(tokenPools.Pools))
		results[testName] = true
	}

	// -------------------------
	// Test SearchService
	// -------------------------
	fmt.Println("\n📝 TESTING SearchService")

	// 1. Search
	testName = "SearchService.Search"
	fmt.Printf("Testing %s... ", testName)
	searchResults, err = client.Search.Search(ctx, "ethereum")
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Found %d tokens, %d pools, %d DEXes\n",
			len(searchResults.Tokens),
			len(searchResults.Pools),
			len(searchResults.Dexes))
		results[testName] = true
	}

	// -------------------------
	// Test UtilsService
	// -------------------------
	fmt.Println("\n📝 TESTING UtilsService")

	// 1. Get stats
	testName = "UtilsService.GetStats"
	fmt.Printf("Testing %s... ", testName)
	stats, err := client.Utils.GetStats(ctx)
	if err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
		results[testName] = false
	} else {
		fmt.Printf("✅ SUCCESS: Stats - Chains: %d, Factories: %d, Pools: %d, Tokens: %d\n",
			stats.Chains,
			stats.Factories,
			stats.Pools,
			stats.Tokens)
		results[testName] = true
	}

	// -------------------------
	// Summary of test results
	// -------------------------
	fmt.Println("\n📊 TEST RESULTS SUMMARY")
	fmt.Println("===============================================")

	totalTests := len(results)
	passedTests := 0

	for test, passed := range results {
		if passed {
			passedTests++
			fmt.Printf("✅ %s\n", test)
		} else {
			fmt.Printf("❌ %s\n", test)
		}
	}

	fmt.Printf("\n%d/%d tests passed (%.1f%%)\n",
		passedTests,
		totalTests,
		float64(passedTests)/float64(totalTests)*100)

	if passedTests == totalTests {
		fmt.Println("\n🎉 All DexPaprika SDK methods are working correctly!")
	} else {
		fmt.Printf("\n⚠️ %d tests failed. Please check the logs above for details.\n", totalTests-passedTests)
	}
}
