package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var downstreamNodes []DownstreamNodeConfig
var ctx, cancel = context.WithCancel(context.Background())
var ctx1, cancel1 = context.WithCancel(context.Background())
var Initflag = false

func init() {
	// 配置 Zerolog：输出到控制台（彩色格式），带时间戳和调用者信息
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Caller().
		Logger()
}

// 设置下游节点时会先停止原先的虚拟链路
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
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

	log.Info().Msg("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func setNodesHandler(w http.ResponseWriter, r *http.Request) {
	ctx1, cancel1 = context.WithCancel(context.Background())
	log.Debug().Msg("Downstreamset 1")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Warn().Str("method", r.Method).Msg("Invalid request method")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Error reading request body")
		return
	}
	if Initflag {
		cancel()
		log.Info().Msg("Cancelled previous traffic generation")
	}

	var nodes []DownstreamNodeConfig
	if err := json.Unmarshal(body, &nodes); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	downstreamNodes = nodes
	log.Info().Interface("nodes", nodes).Msg("Downstream nodes updated")
	fmt.Fprintf(w, "Downstream nodes set la1: %v\n", downstreamNodes)
	Initflag = true
	ctx, cancel = ctx1, cancel1
}

func startTrafficHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("startTrafficHandler 1")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	for _, downstreamNode := range downstreamNodes {
		go TrafficControl(downstreamNode, ctx)
	}
	fmt.Fprintf(w, "Traffic generation started\n")
	fmt.Printf("Traffic generation started\n")
}

// 这里是发送流量的具体逻辑
func TrafficControl(node DownstreamNodeConfig, ctx context.Context) {
	fmt.Println("TrafficControl 1")
	fmt.Printf("TrafficControl Duration %v", 1.0/node.Rps)
	ticker := time.NewTicker(time.Second * time.Duration(1.0/node.Rps))
	defer ticker.Stop()
	// 添加请求超时控制
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	for {
		for i := 1; i <= 100; i++ {
			select {
			case <-ctx.Done():
				fmt.Println("Traffic sending stopped")
				return
			case <-ticker.C:
				if float64(i) <= float64(node.ErrorRate) {
					sendTraffic(client, true, node.DNS)
				} else {
					sendTraffic(client, false, node.DNS)
				}
			}
		}

	}
}

func sendTraffic(client *http.Client, isConfusion bool, nodeName string) error {
	fmt.Println("sendTraffic 1")
	url := nodeName
	fmt.Printf("Url %s\n", url)
	req, err := http.NewRequest("GET", "http://"+url, nil)
	if err != nil {
		fmt.Printf("Error creating request to %s: %v\n", nodeName, err)
		return err
	}
	// 添加固定的请求头
	if !isConfusion {
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
