package main

import (
	"fmt"
)

func main() {
	// 初始化 Kubernetes 客户端
	clientset, err := initKubernetesClient()
	if err != nil {
		panic(err)
	}

	// 配置
	namespace := "default"
	name := "my-service"
	image := "nginx:latest"
	replicas := int32(1)
	port := int32(8080)
	targetPort := int32(80)

	// 创建 Deployment
	err = createDeployment(clientset, namespace, name, image, replicas)
	if err != nil {
		fmt.Printf("Error creating deployment: %v\n", err)
		return
	}

	// 创建 Service
	err = createService(clientset, namespace, name, port, targetPort)
	if err != nil {
		fmt.Printf("Error creating service: %v\n", err)
		return
	}
}
