package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/sideChannel_topo_confusion/conductorEngine/pkg/controller/topo"
)

func main() {
	// // 初始化 Kubernetes 客户端
	// clientset, err := initKubernetesClient()
	// if err != nil {
	// 	panic(err)
	// }

	// // 配置
	// namespace := "default"
	// name := "my-service"
	// image := "nginx:latest"
	// replicas := int32(1)
	// port := int32(8080)
	// targetPort := int32(80)

	// // 创建 Deployment
	// err = createDeployment(clientset, namespace, name, image, replicas)
	// if err != nil {
	// 	fmt.Printf("Error creating deployment: %v\n", err)
	// 	return
	// }

	// // 创建 Service
	// err = createService(clientset, namespace, name, port, targetPort)
	// if err != nil {
	// 	fmt.Printf("Error creating service: %v\n", err)
	// 	return
	// }
	// 查询 istio_requests_total 指标（请求总数）
	// Prometheus API client setup
	// Prometheus API client setup
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:9090",
	})
	if err != nil {
		fmt.Printf("Error creating Prometheus client: %v\n", err)
		return
	}
	promApi := v1.NewAPI(client)

	// Parameters
	namespace := "default"
	queryTime := time.Now()
	duration := 5 * time.Minute

	// Generate traffic map
	trafficMap := topo.GenerateTrafficMap(promApi, namespace, queryTime, duration)

	// Print traffic map
	for id, node := range trafficMap {
		fmt.Printf("Node: %s\n", id)
		for _, edge := range node.Edges {
			fmt.Printf("  Edge to %s, Traffic: %v\n", edge.Dest.ID, edge.Metadata["traffic"])
		}
	}
}
