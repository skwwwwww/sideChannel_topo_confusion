package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var downstreamNodes []string

func main() {
	// 从环境变量读取初始节点
	if envNodes := os.Getenv("INITIAL_NODES"); envNodes != "" {
		if err := json.Unmarshal([]byte(envNodes), &downstreamNodes); err != nil {
			fmt.Printf("Error parsing INITIAL_NODES: %v\n", err)
		}
	}

	http.HandleFunc("/set-nodes", setNodesHandler)
	http.HandleFunc("/start-traffic", startTrafficHandler)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

func setNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var nodes []string
	if err := json.Unmarshal(body, &nodes); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	downstreamNodes = nodes
	fmt.Fprintf(w, "Downstream nodes set: %v\n", downstreamNodes)
}
func startTrafficHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	go sendTraffic()
	fmt.Fprintf(w, "Traffic generation started\n")
}

func sendTraffic() {
	for {
		for _, node := range downstreamNodes {
			url := fmt.Sprintf("http://%s/health", node)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("Error creating request to %s: %v\n", node, err)
				continue
			}

			// 添加固定的请求头
			req.Header.Add("X-Traffic-Type", "confusion")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error sending request to %s: %v\n", node, err)
				continue
			}

			fmt.Printf("Response from %s: %s\n", node, resp.Status)
			resp.Body.Close()
		}

		// 每隔5秒发送一次流量
		time.Sleep(5 * time.Second)
	}
}
