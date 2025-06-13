package generatetg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	// criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
	generateobfucationstrategy "github.com/sideChannel_topo_confusion/ce/generateobfucationstrategy"
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

const (
	maxRetries = 10              // 最大重试次数
	retryDelay = 3 * time.Second // 基础重试间隔
)

var gvr = schema.GroupVersionResource{
	Group:    "networking.istio.io",
	Version:  "v1alpha3",
	Resource: "envoyfilters",
}

// 这里搞一个全局的client，用来创建相关的资源
// 使用之前要先init
var Client dynamic.Interface
var Clientset *kubernetes.Clientset

// client初始化
func InitClient() {
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// 创建 DynamicClient
	Client, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("无法创建 DynamicClient: %v", err)
	}

	// 创建 Kubernetes 客户端
	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
}

// 创建OA
// 返回值
func CreateOA(namespace string, instanceID string) *corev1.Service {
	// 创建 Deployment
	deployment := configDeployment(instanceID)
	_, err := Clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create deployment %s: %v", instanceID, err)
	}
	fmt.Printf("Deployment created successfully: traffic-service-%s\n", instanceID)

	// 创建 Service
	service := configService(instanceID)
	createService, err := Clientset.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Failed to create service %s: %v", instanceID, err)
	}
	fmt.Printf("Service created successfully: traffic-service-%s\n", instanceID)
	return createService
}

// 这里创建EnvoyFilter
func CreateEnvoyFilter(namespace string, instanceName string) {
	envoyConfig := EnvoyFilterConfig{}
	// envoyConfig.App = "traffic-service-" + instanceID
	envoyConfig.App = instanceName
	envoyConfig.Namespace = namespace
	envoyFilter := configEnvoyFilter(envoyConfig, envoyConfig.App)
	// 创建 EnvoyFilter
	_, err := Client.Resource(gvr).Namespace(namespace).Create(context.TODO(), envoyFilter, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("无法创建 EnvoyFilter: %v", err)
	}

	fmt.Printf("成功创建 EnvoyFilter: %s/%s\n", namespace, envoyConfig.App)

}

func CreateRootEnvoyFilter(namespace string, instanceName string) {
	envoyConfig := EnvoyFilterConfig{}
	// envoyConfig.App = "traffic-service-" + instanceID
	envoyConfig.App = instanceName
	envoyConfig.Namespace = namespace
	envoyFilter := configRootEnvoyFilter(envoyConfig, envoyConfig.App)
	// 创建 EnvoyFilter
	_, err := Client.Resource(gvr).Namespace(namespace).Create(context.TODO(), envoyFilter, metav1.CreateOptions{})
	if err != nil {
		fmt.Errorf("无法创建 EnvoyFilter: %v", err)
	}

	fmt.Printf("成功创建 EnvoyFilter: %s/%s\n", namespace, envoyConfig.App)

}

// 这里我觉的还需要解耦以下，他的责任只是创建OA
// func GeneralTrafficGenertator() {

// 	for i := 0; i < instanceCount-1; i++ {

// 		nextNodes := getNextNLayers(hostAndPorts, i)
// 		serviceURL := "http://" + hostAndPorts[i][0] + "/healthz"
// 		for {
// 			err := checkService(serviceURL)
// 			if err == nil {
// 				break
// 			}

// 			if i >= maxRetries {
// 				fmt.Errorf("超过最大重试次数%d次", maxRetries)
// 			}

// 			delay := retryDelay * time.Duration(i+1) // 线性退避
// 			fmt.Printf("第%d次重试，等待%s后重试\n", i+1, delay)
// 			time.Sleep(delay)
// 		}
// 		SetDownstreamNode(nextNodes, hostAndPorts[i][0])

// 	}

// }

// 配置 Deployment
func configDeployment(instanceID string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-service-" + instanceID,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "traffic-service-" + instanceID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "traffic-service-" + instanceID,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "traffic-generator",
							Image: "obfuscation_agent:v3",
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
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
								TimeoutSeconds:      1,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
						},
					},
				},
			},
		},
	}
}

// 配置 Service
func configService(instanceID string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-service-" + instanceID,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "traffic-service-" + instanceID,
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

// 配置 EnvoyFilter
// 这里的配置是为了让EnvoyFilter可以拦截流量
func configEnvoyFilter(config EnvoyFilterConfig, instanceName string) *unstructured.Unstructured {
	// 定义 EnvoyFilter 对象
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "EnvoyFilter",
			"metadata": map[string]interface{}{
				"name":      "filter-confusion-header" + instanceName,
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
  if headers:get("X-Traffic-Type") == "normal" then
    request_handle:respond({[":status"] = "500"}, "Request header not allowed")
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

// 配置 EnvoyFilter
// 这里的配置是为了让EnvoyFilter可以拦截流量
func configRootEnvoyFilter(config EnvoyFilterConfig, dstDNS string) *unstructured.Unstructured {
	// 定义 Lua 脚本模板，转义 % 字符并使用 %s 占位符
	luaTemplate := `
function generate_span_id()
    local file = io.open("/dev/urandom", "rb")
    if not file then return nil end
    local bytes = file:read(8)
    file:close()
    return string.format("%%02x%%02x%%02x%%02x%%02x%%02x%%02x%%02x",
        string.byte(bytes, 1), string.byte(bytes, 2), string.byte(bytes, 3), string.byte(bytes, 4),
        string.byte(bytes, 5), string.byte(bytes, 6), string.byte(bytes, 7), string.byte(bytes, 8))
end

function envoy_on_request(request_handle)
    local original_headers = request_handle:headers()
    local headers_to_send = {
        [":method"] = "GET",
        [":path"] = "/api",
        [":authority"] = "%s",
        [":scheme"] = "http",
        ["X-Traffic-Type"] = "confusion",
        ["Content-Length"] = "0"
    }
    local request_id = original_headers:get("x-request-id")
    if request_id then
        headers_to_send["x-request-id"] = request_id
    end
    local trace_id = original_headers:get("x-b3-traceid")
    if trace_id then
        headers_to_send["x-b3-traceid"] = trace_id
        local parent_span_id = original_headers:get("x-b3-spanid")
        if parent_span_id then
            headers_to_send["x-b3-parentspanid"] = parent_span_id
        end
        headers_to_send["x-b3-spanid"] = generate_span_id()
        local sampled = original_headers:get("x-b3-sampled")
        if sampled then
            headers_to_send["x-b3-sampled"] = sampled
        end
    end
    local istio_attributes = original_headers:get("x-istio-attributes")
    if istio_attributes then
        headers_to_send["x-istio-attributes"] = istio_attributes
    end
    request_handle:logWarn("生成带有跟踪上下文的新请求...")
    local cluster_name = "outbound|80||%s"
    local ok, err = request_handle:httpCall(
        cluster_name,
        headers_to_send,
        nil, -- request body
        5000 -- timeout
    )
    if not ok then
        request_handle:logWarn("生成请求失败: " .. tostring(err))
    else
        request_handle:logWarn("已成功发送带有跟踪上下文的请求。")
    end
end
`

	// 使用 dstDNS 格式化 Lua 脚本
	luaCode := fmt.Sprintf(luaTemplate, dstDNS, dstDNS)

	// 定义 EnvoyFilter 对象
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "EnvoyFilter",
			"metadata": map[string]interface{}{
				"name":      "filter-confusion-header-root",
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
							"context": "SIDECAR_OUTBOUND",
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
									"@type":      "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua",
									"inlineCode": luaCode,
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
	// 删除 EnvoyFilter
	err := client.Resource(gvr).Namespace(config.Namespace).Delete(context.TODO(), "filter-confusion-header", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("无法删除 EnvoyFilter: %v", err)
	}

	fmt.Printf("成功删除 EnvoyFilter: %s/%s\n", config.Namespace, "filter-confusion-header")
	return nil
}

func deleteService(clientset *kubernetes.Clientset, namespace string, instanceID string) error {
	// 删除 Service
	err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), "traffic-service-"+instanceID, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("无法删除 Service: %v", err)
	}

	fmt.Printf("成功删除 Service: %s/%s\n", namespace, "traffic-service-"+instanceID)
	return nil
}

func SetDownstreamNode(downstreamNodeConfigs []generateobfucationstrategy.DownstreamNodeConfig, service string) {
	// urls := []string{}
	for i := 0; i < len(downstreamNodeConfigs); i++ {
		downstreamNodeConfigs[i].DNS += "/api"
	}
	targetSetNode := "http://" + service + "/set-nodes"

	client := &http.Client{}

	// 序列化为 JSON
	requestBody, err := json.Marshal(downstreamNodeConfigs)
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
	}
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", targetSetNode, bytes.NewBuffer(requestBody))
	//这里也要将分开
	req.Header.Set("Host", strings.Split(service, ":")[0])
	if err != nil {
		log.Fatalf("无法创建 HTTP 请求: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 记录请求头
	log.Printf("请求头: %+v", req.Header)
	log.Printf("Url: %+v", req.URL)

	// 记录请求体
	log.Printf("请求体(Raw JSON): %s", string(requestBody))

	//记录service
	log.Printf("Dest Service(目标服务): %+v", targetSetNode)
	// 发送请求
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

	// 记录请求头
	log.Printf("请求头: %+v", req1.Header)
	log.Printf("Url: %+v", req1.URL)

	//client1 := &http.Client{}
	resp, err = client.Do(req1)
	if err != nil {
		log.Fatalf("发送 HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 4. 读取响应
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("读取响应失败: %v", err)
	}

	// 5. 输出响应
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))

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
