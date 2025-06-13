package generateobfucationstrategy

// 这里只需要传入分析出来的关键路径
// 需要关键路径的相关指标
// 跟据路径创建OA，EnvoyFilter

// 创建完 OA和EnvoyFilter后，
// 跟据相关指标设置下游节点
import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
)

// 需要返回三部分
// 1. 关键路径长度即为，OA的数量，配置
// 2. 再哪些Service中创建EnvoyFilter
// 3. 给出每个OA的下游节点，并跟据不同的应用类型设置下游节点的指标

type OAConfig struct {
	Namespace  string
	InstanceID string
}

type EnvoyFilterConfig struct {
	Namespace    string
	InstanceName string
}

type DownstreamNodeConfig struct {
	DNS        string
	ServiceNum int
	Rps        float64
	ErrorRate  float64
}

func Generateconfucationstrategy() ([]OAConfig, []EnvoyFilterConfig, [][]DownstreamNodeConfig) {
	// path中包含了每个关键路径的
	_, path, nodesMap, criticalPathNodeMetrics := criticalpath.GetCriticalPaths()
	oaConfig := GenerateOAStrategy(path, nodesMap)
	envoyFilterConfig := GenerateEFStrategy(path, nodesMap)
	downstreamNodeConfig := GenerateDownstreamNodesStrategy(path, nodesMap, criticalPathNodeMetrics, criticalpath.SERVICE_DEPENDENCY)

	return oaConfig, envoyFilterConfig, downstreamNodeConfig
}

// CreateOA 只需要 namespace string, instanceID string
func GenerateOAStrategy(path []string, nodesMap map[string]criticalpath.TrafficNode) []OAConfig {
	len := len(path)
	oaConfig := make([]OAConfig, len)
	for i := 0; i < len; i++ {
		oaConfig[i].Namespace = nodesMap[path[i]].Namespace
		oaConfig[i].InstanceID = fmt.Sprintf("%d", i)
	}
	return oaConfig
}

// CreateEnvoyFilter 只需要 namespace string, instanceName string
// 这里的Name分为两部分, 一部分是自己创建的ID，一部分是关键路径上的Service
func GenerateEFStrategy(path []string, nodesMap map[string]criticalpath.TrafficNode) []EnvoyFilterConfig {
	len := len(path)
	fmt.Println("path len : %d", len)
	envoyFilterConfig := make([]EnvoyFilterConfig, len*2)

	for i := 0; i < len; i++ {
		envoyFilterConfig[i].Namespace = nodesMap[path[i]].Namespace
		envoyFilterConfig[i].InstanceName = fmt.Sprintf("traffic-service-%d", i)
	}
	for i := 1; i < len; i++ {
		envoyFilterConfig[i+len-1].Namespace = nodesMap[path[i]].Namespace
		envoyFilterConfig[i+len-1].InstanceName = nodesMap[path[i]].App
	}
	if len > 0 {
		envoyFilterConfig[len*2-1].Namespace = nodesMap[path[0]].Namespace
		envoyFilterConfig[len*2-1].InstanceName = nodesMap[path[0]].App
	}
	return envoyFilterConfig
}

// 这里负责给出每个OA的下游节点 以及 OA与每一个下游节点的虚拟链路的指标（RPS， 时延， 错误率）
func GenerateDownstreamNodesStrategy(path []string, nodesMap map[string]criticalpath.TrafficNode, criticalPathNodeMetrics []criticalpath.CriticalPathNodeMetric, ApplicationClass int) [][]DownstreamNodeConfig {
	// 这里需要根据OA的数量来创建
	// 每一个OA都对应了一个下游节点配置
	n := len(path)
	fmt.Printf("critical Path: %v\n", path)
	wService, wRPS, wError := getObfucationWeights(ApplicationClass)

	// 这里需要确定一下每一层的RPS，等相关zhi
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}
	// 创建 Kubernetes 客户端
	Clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	downstreamNodeConfigs := make([][]DownstreamNodeConfig, n)
	for i := 0; i+1 < n; i++ {
		for j := 1; j <= n-2 && j+i < n; j++ {

			// 这个namespace是可以二次利用的
			namespace := nodesMap[path[i]].Namespace
			// i + j 就是下一层到下N-2层
			dnsName := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+fmt.Sprint(j+i), namespace)
			// 先把OA加入进去
			downstreamNodeConfigOA := DownstreamNodeConfig{
				DNS:        dnsName + ":80",
				ServiceNum: criticalPathNodeMetrics[i+1].ServiceNum * wService,
				Rps:        criticalPathNodeMetrics[i+1].Rps * wRPS,
				ErrorRate:  criticalPathNodeMetrics[i+1].ErrorRate * wError,
			}
			downstreamNodeConfigs[i] = append(downstreamNodeConfigs[i], downstreamNodeConfigOA)
			app := nodesMap[path[i+j]].App
			service, err := Clientset.CoreV1().Services(namespace).Get(context.TODO(), app, metav1.GetOptions{})
			if err != nil {
				log.Fatalf("无法获取 Service: %v", err)
			}

			// 6. 提取 ClusterIP (host) 和 Port
			dnsName = fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)

			// 获取第一个端口（如果有多个端口，可以根据需要选择）
			if len(service.Spec.Ports) == 0 {
				log.Fatal("Service 没有定义任何端口")
			}
			port := service.Spec.Ports[0].Port

			downstreamNodeConfig := DownstreamNodeConfig{
				DNS:        "" + dnsName + ":" + fmt.Sprintf("%d", port),
				ServiceNum: criticalPathNodeMetrics[i+1].ServiceNum * wService,
				Rps:        criticalPathNodeMetrics[i+1].Rps * wRPS,
				ErrorRate:  criticalPathNodeMetrics[i+1].ErrorRate * wError,
			}
			downstreamNodeConfigs[i] = append(downstreamNodeConfigs[i], downstreamNodeConfig)
		}
	}
	fmt.Printf("downstreamNodeConfigs : %v\n", downstreamNodeConfigs)
	return downstreamNodeConfigs
}

// getWeights 根据应用类型返回对应的权重
func getObfucationWeights(appType int) (wService int, wRPS, wError float64) {
	switch appType {
	case criticalpath.SERVICE_DEPENDENCY:
		return 1, 0.5, 0.0
	case criticalpath.DELAY_SENSITIVE:
		return 1, 0.5, 0.0
	case criticalpath.TRAFFIC_INTENSIVE:
		return 1, 1.2, 0.0
	case criticalpath.RELIABILITY_PRIORITY:
		return 1, 0.5, 1.2
	case criticalpath.RESOUTRCE_INTENSIVE:
		return 1, 0.5, 1.0
	default:
		return 1, 1.0, 1.0
	}
}
