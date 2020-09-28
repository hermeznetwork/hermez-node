package eth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func addBlock(url string) {
	method := "POST"

	payload := strings.NewReader("{\n    \"jsonrpc\":\"2.0\",\n    \"method\":\"evm_mine\",\n    \"params\":[],\n    \"id\":1\n}")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()
}

func addBlocks(numBlocks int64, url string) {
	for i := int64(0); i < numBlocks; i++ {
		addBlock(url)
	}
}

func addTime(seconds float64, url string) {
	secondsStr := strconv.FormatFloat(seconds, 'E', -1, 32)

	method := "POST"
	payload := strings.NewReader("{\n    \"jsonrpc\":\"2.0\",\n    \"method\":\"evm_increaseTime\",\n    \"params\":[" + secondsStr + "],\n    \"id\":1\n}")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()
}
