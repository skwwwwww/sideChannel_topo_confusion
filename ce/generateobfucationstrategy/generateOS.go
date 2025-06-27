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

func Generateconfucationstrategy(path []string, nodesMap map[string]criticalpath.TrafficNode, criticalPathNodeMetrics []criticalpath.CriticalPathNodeMetric) ([]OAConfig, []EnvoyFilterConfig, [][]DownstreamNodeConfig) {
	// path中包含了每个关键路径的
	// _, path, nodesMap, criticalPathNodeMetrics := criticalpath.GetCriticalPaths()
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
	log.Printf("path len : %d\n", len)
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
	log.Printf("critical Path: %v\n", path)
	wService, wRPS, wError := getObfucationWeights(ApplicationClass)

	// 这里需要确定一下每一层的RPS，等相关zhi
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		log.Fatalf("无法加载kubeconfig for GenerateDownstreamNodesStrategy: %v", err)
	}
	// 创建 Kubernetes 客户端
	Clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("无法创建Kubernetes客户端 for GenerateDownstreamNodesStrategy: %v", err)
	}

	downstreamNodeConfigs := make([][]DownstreamNodeConfig, n)
	// 遍历每个OA实例（从0到 n-2），因为最后一层OA不需要设置下游节点
	for i := 0; i < n; i++ {
		if i+1 >= n { // 最后一层OA不需要设置下游节点
			continue
		}

		// `j` 代表从当前OA (i) 到下游节点在关键路径中的相对距离
		// 这意味着每个OA[i]会连接到 OA[i+1], OA[i+2], ..., OA[n-1]
		// 以及对应的 App[i+1], App[i+2], ..., App[n-1]
		for j := 1; i+j < n; j++ {
			targetPathNodeIndex := i + j // 目标节点在关键路径中的索引

			// 配置指向另一个 OA 实例的下游节点
			targetOANamespace := nodesMap[path[targetPathNodeIndex]].Namespace
			targetOAInstanceID := fmt.Sprintf("%d", targetPathNodeIndex)
			oaDNS := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+targetOAInstanceID, targetOANamespace)

			downstreamNodeConfigOA := DownstreamNodeConfig{
				DNS:        oaDNS + ":80", // 假设 traffic-service 监听 80 端口
				ServiceNum: criticalPathNodeMetrics[targetPathNodeIndex].ServiceNum * wService,
				Rps:        criticalPathNodeMetrics[targetPathNodeIndex].Rps * wRPS,
				ErrorRate:  criticalPathNodeMetrics[targetPathNodeIndex].ErrorRate * wError,
			}
			downstreamNodeConfigs[i] = append(downstreamNodeConfigs[i], downstreamNodeConfigOA)

			// 配置指向原始应用的下游节点
			app := nodesMap[path[targetPathNodeIndex]].App
			appNamespace := nodesMap[path[targetPathNodeIndex]].Namespace // 应用的命名空间
			service, err := Clientset.CoreV1().Services(appNamespace).Get(context.TODO(), app, metav1.GetOptions{})
			if err != nil {
				log.Printf("警告: 无法获取服务 %s 在命名空间 %s 中: %v。跳过此下游节点配置。\n", app, appNamespace, err)
				continue // 如果服务不存在，则跳过此下游节点
			}

			if len(service.Spec.Ports) == 0 {
				log.Printf("警告: 服务 %s 在命名空间 %s 中未定义任何端口。跳过此下游节点配置。\n", app, appNamespace)
				continue
			}
			port := service.Spec.Ports[0].Port

			appDNS := fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
			downstreamNodeConfigApp := DownstreamNodeConfig{
				DNS:        appDNS + ":" + fmt.Sprintf("%d", port),
				ServiceNum: criticalPathNodeMetrics[targetPathNodeIndex].ServiceNum * wService,
				Rps:        criticalPathNodeMetrics[targetPathNodeIndex].Rps * wRPS,
				ErrorRate:  criticalPathNodeMetrics[targetPathNodeIndex].ErrorRate * wError,
			}
			downstreamNodeConfigs[i] = append(downstreamNodeConfigs[i], downstreamNodeConfigApp)
		}
	}
	fmt.Printf("下游节点配置: %v\n", downstreamNodeConfigs)
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
