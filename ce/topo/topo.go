package topo

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Root 结构体
type Root struct {
	Timestamp int      `json:"timestamp"`
	Duration  int      `json:"duration"`
	GraphType string   `json:"graphType"`
	Elements  Elements `json:"elements"`
}

// Elements 结构体
type Elements struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

var (
	maxRetries = 100             // 重试次数
	retryDelay = 2 * time.Second // 基础重试间隔
)

// 目前Url -> Json的比较完整了
func GetTopo() Root {
	test_url1 := "http://192.168.88.151:20001/kiali/api/namespaces/graph?duration=60s&graphType=service&includeIdleEdges=false&injectServiceNodes=true&boxBy=cluster,namespace&ambientTraffic=none&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health,idleNode&rateGrpc=requests&rateHttp=requests&rateTcp=none&namespaces=default"
	// test_url1 := "http://kiali.istio-system.svc.cluster.local:20001/kiali/api/namespaces/graph?duration=60s&graphType=service&includeIdleEdges=false&injectServiceNodes=true&boxBy=cluster,namespace&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=default,istio-system,local-path-storage,sockshop-coherence,sockshop-core"
	resp, err := http.Get(test_url1)
	if resp.StatusCode != 200 {
		numRetries := 0
		for {
			resp, err = http.Get(test_url1)
			if numRetries >= maxRetries {
				break // 超过最大重试次数，退出重试循环
			}
			if err == nil && resp.StatusCode == 200 {
				log.Println("获取拓扑成功")
				break // 成功获取数据，退出重试循环
			}
			// 如果失败，进行线性退避重试
			delay := retryDelay * time.Duration(numRetries+1) // 线性退避
			log.Printf("获取拓扑失败，正在第 %d 次重试，等待 %s...\n", numRetries+1, delay)
			time.Sleep(delay)
			numRetries++
		}
	}

	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	log.Println("获取到的拓扑数据:", resp)
	// 将响应的 JSON 数据解析到结构体
	var root Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return root
}
