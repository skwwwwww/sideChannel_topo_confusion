package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var downstreamNodes []NodeInfo
var stopChanList = []chan struct{}{}
func main() {
	// 从环境变量读取初始节点
	// 这个等大后期再研究
	if envNodes := os.Getenv("INITIAL_NODES"); envNodes != "" {
		if err := json.Unmarshal([]byte(envNodes), &downstreamNodes); err != nil {
			fmt.Printf("Error parsing INITIAL_NODES: %v\n", err)
		}
	}

	http.HandleFunc("/set-nodes", setNodesHandler)
	http.HandleFunc("/start-traffic", startTrafficHandler)
	http.HandleFunc("/stop-traffic", stopTrafficHandler)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

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

	var nodes []NodeInfo
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

	for _, downstreamNode := range downstreamNodes {

		stopChan := make(chan struct{})
		stopChanList = append(stopChanList,stopChan)
		go TrafficControl(downstreamNode,stopChan)
	}
	fmt.Fprintf(w, "Traffic generation started\n")
}

func stopTrafficHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	for _, stopChan := range stopChanList {	
		close(stopChan)
	}
	fmt.Fprintf(w, "Traffic generation stoped\n")
}

// 这里是发送流量的具体逻辑
func TrafficControl(node NodeInfo, stop <-chan struct{}) {
	ticker := time.NewTicker(time.Second * time.Duration(node.nodeRPS))
	defer ticker.Stop()
	// 添加请求超时控制
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	for {
		for i := 1; i <= 100; i++{
			select {
			case <-stop:
				fmt.Println("Traffic sending stopped")
				return
			case <-ticker.C:
				if i <= node.nodeErrorRate {
					sendTraffic(client,true,node.nodeName)	
				} else {
					sendTraffic(client,false,node.nodeName)
				}
			}
		}
		
	}
}

func sendTraffic(client *http.Client,isConfusion bool,nodeName string) (error){
	url := fmt.Sprintf("http://%s", nodeName)
	fmt.Printf("Url %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request to %s: %v\n", nodeName, err)
		return err
	}
	// 添加固定的请求头	
	if isConfusion {
		req.Header.Add("X-Traffic-Type", "confusion")
	} else {
		req.Header.Add("X-Traffic-Type", "normal")
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request to %s: %v\n", nodeName, err)
		return err
	}
	fmt.Printf("Response from %s: %s\n", nodeName, resp.Status)
	resp.Body.Close()
	return nil
}

// HealthStatus represents the health status of the application.
type HealthStatus struct {
	Status string `json:"status"`
}

// healthz handles the health check request.
func healthz(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查应用程序的健康状态
	if isHealthy() {
		// 如果健康，返回200 OK
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthStatus{Status: "healthy"})
	} else {
		// 如果不健康，返回503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthStatus{Status: "unhealthy"})
	}
}

// ReadinessStatus represents the readiness status of the application.
type ReadinessStatus struct {
	Status string `json:"status"`
}

// ready handles the readiness check request.
func ready(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查应用程序的就绪状态
	if isReady() {
		// 如果就绪，返回200 OK
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ReadinessStatus{Status: "ready"})
	} else {
		// 如果未就绪，返回503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ReadinessStatus{Status: "not_ready"})
	}
}

// isReady checks if the application is ready.
// Replace this with your actual readiness check logic.
func isReady() bool {
	// 示例：检查缓存是否已加载
	// return cache.IsLoaded()

	// 这里只是一个示例，返回true表示就绪
	return true
}

// isHealthy checks if the application is healthy.
// Replace this with your actual health check logic.
func isHealthy() bool {
	// 示例：检查数据库连接
	// return db.Ping() == nil

	// 这里只是一个示例，返回true表示健康
	return true
}
