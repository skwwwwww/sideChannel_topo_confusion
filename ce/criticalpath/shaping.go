package criticalpath

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

const (
	SERVICE_DEPENDENCY   = 1
	DELAY_SENSITIVE      = 2
	TRAFFIC_INTENSIVE    = 3
	RELIABILITY_PRIORITY = 4
	RESOUTRCE_INTENSIVE  = 5

	TRAFFIC = "traffic"
	METRICS = "metrics"
)

type Topo struct {
	Nodes map[string]topo.Node
	Edges map[string]topo.Edge
}

type TrafficNode struct {
	num         int
	namespace   string
	serviceName string
}

type TrafficEdge struct {
	isConnected bool
	serviceNum  int
	rps         float64
	errorRate   float64
	trafficSize int
}

func Shaped(root topo.Root) Topo {
	topo := Topo{
		Nodes: make(map[string]topo.Node),
		Edges: make(map[string]topo.Edge),
	}
	for _, v := range root.Elements.Nodes {
		topo.Nodes[v.Data.ID] = v
	}
	for _, v := range root.Elements.Edges {
		topo.Edges[v.Data.ID] = v
	}
	return topo
}

// 将从kiali获取到的topo重塑成加权有向连接矩阵
func shaped1(root topo.Root) (trafficMap [][]TrafficEdge, nodes []string, nodesMap map[string]TrafficNode) {
	num := 0
	//获取节点的namespace，serviceName, 以及在图中的顺序
	for i, v := range root.Elements.Nodes {
		nodes = append(nodes, v.Data.ID)
		trafficNode := TrafficNode{
			num:         i,
			namespace:   v.Data.Namespace,
			serviceName: v.Data.Service,
		}
		nodesMap[nodes[i]] = trafficNode
		num++
	}

	trafficMap = make([][]TrafficEdge, num)
	for i := range trafficMap {
		trafficMap[i] = make([]TrafficEdge, num) // 布尔字段自动初始化为 false
	}
	// 遍历边，将每条边的需要用到的值放入其中
	for _, v := range root.Elements.Edges {
		target := v.Data.Target
		source := v.Data.Source
		numTraget := nodesMap[target].num
		numSource := nodesMap[source].num
		trafficEdge := trafficMap[numSource][numTraget]
		trafficEdge.isConnected = true
		// 服务依赖性 -> 边的权值为 "1"
		trafficEdge.serviceNum = 1
		// 流量密集型 -> 边的权值为服务间的流量rps
		trafficEdge.rps, _ = strconv.ParseFloat(v.Data.Traffic.Rates.Http, 64)
		// 可靠优先型 -> 边的权值为错误率
		trafficEdge.errorRate = 100 - v.Data.Traffic.Responses.Flags["-"]
		// 流量密集型 -> 边的权值为服务间的流量大小（这个不知道kiali中能不能获取到）
		//trafficEdge.trafficSize = int(GetValue())
		//topo.Edges[v.Data.ID] = v

	}

	return [][]TrafficEdge{}, []string{}, nodesMap
}

// 根据不同的应用类型返回不同的权值
// 服务依赖性 -> 边的权值为 "1"
// 延迟敏感型 -> 边的权值为服务间的平均延迟
// 流量密集型 -> 边的权值为服务间的流量大小（这个不知道kiali中能不能获取到）
// 可靠优先型 -> 边的权值为错误率
// 资源密集型 -> 待定
func GetValue(applicationType int, namespace string, service string) float64 {

	if applicationType == SERVICE_DEPENDENCY {
		return 1.0
	}
	if applicationType == DELAY_SENSITIVE {

	}
	//先暂定使用rps
	if applicationType == TRAFFIC_INTENSIVE {

	}
	if applicationType == RELIABILITY_PRIORITY {

	}
	return 1.0
}

// metricTpye 目前有两种TRAFFIC和METRICS
func GetTrafficMertics(namespace, service, metricTpye string) {
	url := "http://192.168.200.153:20001/kiali/api/namespaces/" + namespace + "/services/" + service + "/graph?duration=60s&graphType=workload&includeIdleEdges=false&injectServiceNodes=true&responseTime=95&throughputType=request&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,securityPolicy,responseTime,throughput&rateGrpc=requests&rateHttp=requests&rateTcp=sent"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	defer resp.Body.Close()
	//读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	log.Print("%+v", string(body))
	if metricTpye == METRICS {
		var metrics InboundMetrics
		if err := json.Unmarshal(body, &metrics); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v", metrics)
	}

	if metricTpye == TRAFFIC {
		var topology ServiceTopology
		if err := json.Unmarshal(body, &topology); err != nil {
			log.Fatal(err)
		}
		// 访问数据示例
		fmt.Println("第一个节点的应用名称:", topology.Elements.Nodes[0].Data.App)
		fmt.Println("第一条边的协议:", topology.Elements.Edges[0].Data.Traffic.Protocol)
	}

}

// 深度优先搜索算法找到权值和最大的路径
func FindMaxWeightPath(trafficMap [][]TrafficEdge, nodes []string, nodesMap map[string]TrafficNode, applicationType int) (maxWeight float64, maxPath []string) {
	visited := make([]bool, len(nodes))
	var currentPath []string
	var dfs func(node int, currentWeight float64)

	dfs = func(node int, currentWeight float64) {
		visited[node] = true
		currentPath = append(currentPath, nodes[node])

		isLeaf := true
		for i, edge := range trafficMap[node] {
			if edge.isConnected && !visited[i] {
				isLeaf = false
				dfs(i, currentWeight+GetEdgeValue(edge, applicationType))
			}
		}

		if isLeaf && currentWeight > maxWeight {
			maxWeight = currentWeight
			maxPath = append([]string(nil), currentPath...)
		}

		currentPath = currentPath[:len(currentPath)-1]
		visited[node] = false
	}

	for i := range nodes {
		dfs(i, 0)
	}

	return maxWeight, maxPath
}
