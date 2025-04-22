package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Package struct {
	Weight  float64 `json:"weight"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Address string  `json:"address"`
}

var (
	countries = []string{
		"Albania", "Brazil", "Canada", "Denmark", "Egypt",
		"France", "Germany", "Hungary", "India", "Japan",
		"Kenya", "Latvia", "Mexico", "Norway", "Oman",
		"Peru", "Qatar", "Russia", "Sweden", "Turkey",
		"Ukraine", "Vietnam", "Yemen", "Zambia", "Zimbabwe",
	}
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
	printLock sync.Mutex
	wg        sync.WaitGroup
)

func randomCountry() string {
	return countries[rand.Intn(len(countries))]
}

func randomWeight() float64 {
	return 0.5 + rand.Float64()*99.5
}

func randomAddress() string {
	return fmt.Sprintf("%d Some Street", rand.Intn(1000)+1) // +1 чтобы избежать 0
}

func sendRequest(i int) {
	defer wg.Done()

	pkg := Package{
		Weight:  randomWeight(),
		From:    randomCountry(),
		To:      randomCountry(),
		Address: randomAddress(),
	}

	data, err := json.Marshal(pkg)
	if err != nil {
		printLock.Lock()
		fmt.Printf("Request #%d: JSON marshal error: %v\n", i+1, err)
		printLock.Unlock()
		return
	}

	resp, err := client.Post("http://localhost:8121/calculate", "application/json", bytes.NewBuffer(data))
	if err != nil {
		printLock.Lock()
		fmt.Printf("Request #%d: Request failed: %v\n", i+1, err)
		printLock.Unlock()
		return
	}
	defer resp.Body.Close()

	printLock.Lock()
	defer printLock.Unlock()
	fmt.Printf("Request #%d => Status: %s (From: %s, To: %s, Weight: %.2fkg)\n",
		i+1, resp.Status, pkg.From, pkg.To, pkg.Weight)
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	start := time.Now()
	wg.Add(50)

	sem := make(chan struct{}, 10)

	for i := 0; i < 50; i++ {
		sem <- struct{}{}
		go func(idx int) {
			defer func() { <-sem }()
			sendRequest(idx)
		}(i)
	}

	wg.Wait()
	close(sem)

	fmt.Printf("\nAll requests completed in %.2f seconds\n", time.Since(start).Seconds())
}
