package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var downstreamNodes []DownstreamNodeConfig

// TrafficManager 负责管理流量生成协程的生命周期
type TrafficManager struct {
	mu              sync.Mutex             // 保护共享状态
	currentCtx      context.Context        // 当前的上下文
	currentCancel   context.CancelFunc     // 当前的取消函数
	downstreamNodes []DownstreamNodeConfig // 当前的下游节点配置
}

var trafficManager *TrafficManager

// NewTrafficManager 创建并初始化 TrafficManager
func NewTrafficManager() *TrafficManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TrafficManager{
		currentCtx:      ctx,
		currentCancel:   cancel,
		downstreamNodes: make([]DownstreamNodeConfig, 0),
	}
}

// 全局共享的 HTTP 客户端
var sharedHTTPClient *http.Client

// UpdateNodesAndRestartTraffic 更新节点配置并重新启动流量生成
func (tm *TrafficManager) UpdateNodesAndRestartTraffic(newNodes []DownstreamNodeConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 1. 停止之前的协程
	if tm.currentCancel != nil {
		log.Info().Msg("Calling previous cancel function to stop old goroutines.")
		tm.currentCancel()
		// 额外等待一下，确保旧协程有时间退出，避免新的流量立即启动导致瞬时RPS过高
		// 在生产环境中可能需要更复杂的机制来确保所有协程都已停止 (例如 WaitGroup)
		time.Sleep(100 * time.Millisecond)
	}

	// 2. 创建新的 Context 和 CancelFunc
	newCtx, newCancel := context.WithCancel(context.Background())
	tm.currentCtx = newCtx
	tm.currentCancel = newCancel
	tm.downstreamNodes = newNodes

	log.Info().Interface("nodes", tm.downstreamNodes).Msg("Starting new traffic goroutines.")
	// 3. 根据新的配置启动新的协程
	for _, node := range tm.downstreamNodes {
		// 将新的上下文传递给每个 TrafficControl 协程
		go TrafficControl(tm.currentCtx, node)
	}
}

func main() {
	trafficManager = NewTrafficManager()
	// 初始化全局共享的 HTTP 客户端
	sharedHTTPClient = &http.Client{
		Timeout: 10 * time.Second, // 设置全局超时
		// 还可以根据需求配置 Transport，例如调整最大空闲连接数
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			MaxIdleConnsPerHost: 20,               // 每个Host的最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
		},
	}
	http.HandleFunc("/set-nodes", setNodesHandler)

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

	log.Info().Msg("Server started at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func setNodesHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Downstreamset initiated.")

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

	var nodes []DownstreamNodeConfig
	if err := json.Unmarshal(body, &nodes); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		log.Error().Err(err).Msg("Error parsing request body")
		return
	}

	// 使用 TrafficManager 来更新节点并重启流量
	trafficManager.UpdateNodesAndRestartTraffic(nodes)

	log.Info().Interface("nodes", nodes).Msg("Downstream nodes updated and traffic restarted.")
	fmt.Fprintf(w, "Downstream nodes set and traffic restarted: %v\n", nodes)
}

// TrafficControl 是发送流量的具体逻辑，现在接受一个 context 参数
func TrafficControl(ctx context.Context, node DownstreamNodeConfig) {
	log.Info().Str("node_dns", node.DNS).Float64("rps", node.Rps).Float64("error_rate", node.ErrorRate).Msg("TrafficControl goroutine started.")

	// 计算发送间隔
	interval := time.Duration(float64(time.Second) / node.Rps)
	if interval <= 0 { // 防止除以零或非常小的RPS导致无效间隔
		interval = time.Millisecond * 10 // 至少10ms间隔，避免CPU飙升
		log.Warn().Str("node_dns", node.DNS).Msg("RPS is too high or zero, setting interval to 10ms.")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 添加请求超时控制
	client := sharedHTTPClient

	randSource := rand.New(rand.NewSource(time.Now().UnixNano())) // 每个协程使用独立的随机源

	for {
		select {
		case <-ctx.Done():
			// 接收到取消信号，协程优雅退出
			log.Info().Str("node_dns", node.DNS).Msg("TrafficControl goroutine stopping due to context cancellation.")
			return
		case <-ticker.C:
			// 每当 ticker 触发时发送一次流量
			isConfusion := false // 默认不是混淆流量
			// 根据 ErrorRate 决定是否发送混淆流量
			// ErrorRate 假定为百分比，例如 5 代表 5%
			if randSource.Float64()*100 < node.ErrorRate {
				isConfusion = true
			}
			sendTraffic(client, isConfusion, node.DNS)
		}
	}
}

func sendTraffic(client *http.Client, isConfusion bool, nodeName string) error {
	url := "http://" + nodeName
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Error creating request")
		return err
	}

	trafficType := "normal"
	if isConfusion {
		trafficType = "confusion"
	}
	req.Header.Add("X-Traffic-Type", trafficType)

	log.Debug().Str("url", url).Str("traffic_type", trafficType).Msg("Sending traffic")
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("Error sending request")
		return err
	}
	defer resp.Body.Close()
	log.Info().Str("url", url).Str("status", resp.Status).Msg("Response received")
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

func isReady() bool {
	return true
}

func isHealthy() bool {
	return true
}
