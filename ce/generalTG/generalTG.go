package generaltg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type EnvoyFilterConfig struct {
	Namespace string
	App       string
}

func GeneralTrafficGenertator() {
	// 从上一个包获取关键路径
	topo, _, keyPaths, maxDegree := criticalpath.GetCriticalPaths()
	hostAndPorts := [][]string{}

	// 加载 kubeconfig 文件（默认路径为 ~/.kube/config）
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// 创建 DynamicClient
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("无法创建 DynamicClient: %v", err)
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	instanceCount := 0

	// 定义要创建的实例个数,关键节点的度或者关键路径的长度
	maxLen := 0
	for _, v := range keyPaths {
		maxLen = max(len(v.Nodes), maxLen)
	}
	instanceCount = max(maxLen, maxDegree)
	namespace := topo.Nodes[keyPaths[0].Nodes[0]].Data.Namespace

	// 定义 EnvoyFilter 的 GVR (Group Version Resource)
	gvr := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "envoyfilters",
	}
	// 为每个实例创建 Deployment 和 Service
	for i := 1; i <= instanceCount; i++ {

		instanceID := fmt.Sprintf("%d", i) // 生成唯一的实例 ID，例如 a, b, c

		// 创建 Deployment
		deployment := createDeployment(instanceID)
		_, err = clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Failed to create deployment %s: %v", instanceID, err)
			continue
		}
		fmt.Printf("Deployment created successfully: traffic-generator-%s\n", instanceID)

		// 创建 Service
		service := createService(instanceID)
		createService, err := clientset.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Failed to create service %s: %v", instanceID, err)
			continue
		}
		fmt.Printf("Service created successfully: traffic-service-%s\n", instanceID)

		envoyConfig := EnvoyFilterConfig{}
		envoyConfig.App = "traffic-generator-" + instanceID
		envoyConfig.Namespace = namespace
		envoyFilter := createEnvoyFilter(envoyConfig, instanceID)
		// 创建 EnvoyFilter
		_, err = client.Resource(gvr).Namespace(namespace).Create(context.TODO(), envoyFilter, metav1.CreateOptions{})
		if err != nil {
			fmt.Errorf("无法创建 EnvoyFilter: %v", err)
		}

		fmt.Printf("成功创建 EnvoyFilter: %s/%s\n", namespace, "filter-confusion-header")

		// 提取 ClusterIP (host) 和 Port
		host := createService.Spec.ClusterIP
		if host == "" {
			log.Fatal("Service 没有 ClusterIP(可能是 Headless Service)")
		}

		// 获取第一个端口（如果有多个端口，可以根据需要选择）
		if len(createService.Spec.Ports) == 0 {
			log.Fatal("Service 没有定义任何端口")
		}
		port := createService.Spec.Ports[0].Port
		//dnsName := fmt.Sprintf("%s.%s.svc.cluster.local", createService.Name, createService.Namespace)
		//dnsName := fmt.Sprintf("%s.%s.svc", createService.Name, createService.Namespace)
		dnsName := createService.Spec.ClusterIP
		hostAndPort := []string{}
		hostAndPort = append(hostAndPort, ""+dnsName+":"+fmt.Sprint(port))
		hostAndPorts = append(hostAndPorts, hostAndPort)
	}

	for i, v := range keyPaths {
		instanceID := fmt.Sprintf("S%d", i)
		keyPathLen := len(v.Nodes)
		node := v.Nodes[keyPathLen-1]
		namespace := topo.Nodes[node].Data.Namespace
		app := topo.Nodes[node].Data.App
		envoyConfig := EnvoyFilterConfig{}
		envoyConfig.Namespace = namespace
		envoyConfig.App = app

		envoyFilter := createEnvoyFilter(envoyConfig, instanceID)
		// 创建 EnvoyFilter
		_, err := client.Resource(gvr).Namespace(namespace).Create(context.TODO(), envoyFilter, metav1.CreateOptions{})
		if err != nil {
			fmt.Errorf("无法创建 EnvoyFilter: %v", err)
		}

		fmt.Printf("成功创建 EnvoyFilter: %s/%s\n", namespace, "filter-confusion-header")
		service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), app, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("无法获取 Service: %v", err)
		}

		// 6. 提取 ClusterIP (host) 和 Port
		host := service.Spec.ClusterIP
		dnsName := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
		if host == "" {
			log.Fatal("Service 没有 ClusterIP（可能是 Headless Service）")
		}

		// 获取第一个端口（如果有多个端口，可以根据需要选择）
		if len(service.Spec.Ports) == 0 {
			log.Fatal("Service 没有定义任何端口")
		}
		port := service.Spec.Ports[0].Port
		len := len(hostAndPorts)
		hostAndPorts[len-1] = append(hostAndPorts[len-1], ""+dnsName+":"+fmt.Sprintf("%d", port))

	}

	for i := 0; i < instanceCount-1; i++ {

		nextNodes := getNextNLayers(hostAndPorts, i)
		setDownstreamNode(nextNodes, hostAndPorts[i][0])

	}

}

// 创建 Deployment
func createDeployment(instanceID string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-generator-" + instanceID,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "traffic-generator-" + instanceID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "traffic-generator-" + instanceID,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "traffic-generator",
							Image: "traffic-generator:1.0",
							Env: []corev1.EnvVar{
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}
}

// 创建 Service
func createService(instanceID string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-service-" + instanceID,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "traffic-generator-" + instanceID,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func createEnvoyFilter(config EnvoyFilterConfig, instanceID string) *unstructured.Unstructured {
	// 定义 EnvoyFilter 对象
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "EnvoyFilter",
			"metadata": map[string]interface{}{
				"name":      "filter-confusion-header" + instanceID,
				"namespace": config.Namespace,
			},
			"spec": map[string]interface{}{
				"workloadSelector": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": config.App,
					},
				},
				"configPatches": []interface{}{
					map[string]interface{}{
						"applyTo": "HTTP_FILTER",
						"match": map[string]interface{}{
							"context": "SIDECAR_INBOUND",
							"listener": map[string]interface{}{
								"filterChain": map[string]interface{}{
									"filter": map[string]interface{}{
										"name": "envoy.filters.network.http_connection_manager",
										"subFilter": map[string]interface{}{
											"name": "envoy.filters.http.router",
										},
									},
								},
							},
						},
						"patch": map[string]interface{}{
							"operation": "INSERT_BEFORE",
							"value": map[string]interface{}{
								"name": "envoy.filters.http.lua",
								"typed_config": map[string]interface{}{
									"@type": "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua",
									"inlineCode": `
function envoy_on_request(request_handle)
  local headers = request_handle:headers()
  if headers:get("X-Traffic-Type") == "confusion" then
    request_handle:respond({[":status"] = "200"}, "Request header not allowed")
  end
end`,
								},
							},
						},
					},
				},
			},
		},
	}
}

func deleteEnvoyFilter(client dynamic.Interface, config EnvoyFilterConfig) error {
	// 定义 EnvoyFilter 的 GVR (Group Version Resource)
	gvr := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "envoyfilters",
	}

	// 删除 EnvoyFilter
	err := client.Resource(gvr).Namespace(config.Namespace).Delete(context.TODO(), "filter-confusion-header", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("无法删除 EnvoyFilter: %v", err)
	}

	fmt.Printf("成功删除 EnvoyFilter: %s/%s\n", config.Namespace, "filter-confusion-header")
	return nil
}

func setDownstreamNode(nodes []string, service string) {
	urls := []string{}
	for _, v := range nodes {
		urls = append(urls, v+"/api")
	}
	// 序列化为 JSON
	requestBody, err := json.Marshal(urls)
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
	}
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "http://"+service+"/set-nodes", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("无法创建 HTTP 请求: %v", err)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 记录请求头
	log.Printf("请求头: %+v", req.Header)

	// 记录请求体
	log.Printf("Request Body (Raw JSON): %s", string(requestBody))
	log.Printf("请求体: %+v", urls)

	//记录service
	log.Printf("Dest Service(目标服务): %+v", "http://"+service+"/set-nodes")
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("发送 HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应失败: %v", err)
	}

	// 输出响应
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))

	url := "http://" + service + "/start-traffic"
	req1, err1 := http.NewRequest("POST", url, nil)
	if err1 != nil {
		log.Fatalf("无法创建 HTTP 请求: %v", err)
	}
	client1 := &http.Client{}
	resp1, err := client1.Do(req1)
	if err != nil {
		log.Fatalf("发送 HTTP 请求失败: %v", err)
	}
	defer resp1.Body.Close()

	// 4. 读取响应
	body1, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应失败: %v", err)
	}

	// 5. 输出响应
	fmt.Printf("响应状态码: %d\n", resp1.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body1))
}
func getNextNLayers(hostAndPorts [][]string, N int) []string {
	var result []string

	len := len(hostAndPorts)
	for j := 1; j <= len-2 && j+N < len; j++ {
		result = append(result, hostAndPorts[N+j]...)
	}

	return result
}

// 辅助函数：返回 int32 指针
func int32Ptr(i int32) *int32 {
	return &i
}
