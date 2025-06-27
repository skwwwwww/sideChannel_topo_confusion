// main.go

package main

import (
	"encoding/json"
	"fmt"

	"net/http"

	execobfuscationstrategy "github.com/sideChannel_topo_confusion/ce/execobfuscationstrategy"
)

func main() {

	execobfuscationstrategy.Execobfuscationstrategy()

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
	for {
	}
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
