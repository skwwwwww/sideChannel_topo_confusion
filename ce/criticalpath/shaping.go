package criticalpath

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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
	Namespace   string
	ServiceName string
	App         string
}

type TrafficEdge struct {
	isConnected bool
	serviceNum  int
	rps         float64
	errorRate   float64
}

type computeEdge struct {
	to     int
	weight float64
}

// getWeights 根据应用类型返回对应的权重
func getWeights(appType int) (wService, wRPS, wError float64) {
	switch appType {
	case SERVICE_DEPENDENCY:
		return 1.0, 0.0, 0.0
	case DELAY_SENSITIVE:
		return 0.0, -0.5, -0.5
	case TRAFFIC_INTENSIVE:
		return 0.0, 1.0, 0.0
	case RELIABILITY_PRIORITY:
		return 0.0, 0.0, -1.0
	case RESOUTRCE_INTENSIVE:
		return -1.0, 0.0, 0.0
	default:
		return 0.0, 0.0, 0.0
	}
}

type CriticalPathNodeMetric struct {
	ServiceNum int
	Rps        float64
	ErrorRate  float64
}

// 将从kiali获取到的topo重塑成加权有向连接矩阵
// 还得把根节点过滤出来
func shaped1(root topo.Root) (trafficMap [][]TrafficEdge, nodes []string, nodesMap map[string]TrafficNode, rootNodes []TrafficNode) {
	rootNodes = []TrafficNode{}
	num := 0
	nodesMap = make(map[string]TrafficNode)
	//获取节点的namespace，serviceName, 以及在图中的顺序
	for i, v := range root.Elements.Nodes {
		nodes = append(nodes, v.Data.ID)
		trafficNode := TrafficNode{
			num:         i,
			Namespace:   v.Data.Namespace,
			ServiceName: v.Data.Service,
			App:         v.Data.App,
		}
		nodesMap[nodes[i]] = trafficNode
		num++
		if v.Data.IsRoot {
			rootNodes = append(rootNodes, trafficNode)
		}
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
		fmt.Println("target: %s, source: %s, numTraget: %d, numSource: %d", target, source, numTraget, numSource)

		trafficMap[numSource][numTraget].isConnected = true
		// 服务依赖性 -> 边的权值为 "1"
		trafficMap[numSource][numTraget].serviceNum = 1
		// 流量密集型 -> 边的权值为服务间的流量rps
		trafficMap[numSource][numTraget].rps, _ = strconv.ParseFloat(v.Data.Traffic.Rates.Http, 64)
		// 可靠优先型 -> 边的权值为错误率
		if v.Data.Traffic.Responses.Flags["-"] != 0 {
			trafficMap[numSource][numTraget].errorRate = 100 - v.Data.Traffic.Responses.Flags["-"]
		} else {
			trafficMap[numSource][numTraget].errorRate = 0
		}
		fmt.Println(trafficMap[numSource][numTraget])

	}
	fmt.Println(trafficMap)
	return trafficMap, nodes, nodesMap, rootNodes
}

// topologicalSort 使用Kahn算法进行拓扑排序
func topologicalSort(adj [][]computeEdge) []int {
	n := len(adj)
	inDegree := make([]int, n)
	for u := 0; u < n; u++ {
		for _, e := range adj[u] {
			inDegree[e.to]++
		}
	}

	queue := []int{}
	for u := 0; u < n; u++ {
		if inDegree[u] == 0 {
			queue = append(queue, u)
		}
	}

	order := []int{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)
		for _, e := range adj[u] {
			v := e.to
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	return order
}

// FindMaxPath 寻找指标和最大的路径
func FindMaxPath(trafficMap [][]TrafficEdge, nodes []string, nodesMap map[string]TrafficNode, rootNodes []TrafficNode, appType int) (maxSum float64, path []string, criticalPathNodeMetrics []CriticalPathNodeMetric) {

	fmt.Println(trafficMap)

	if len(rootNodes) == 0 || len(nodes) == 0 {
		return 0, nil, nil
	}

	wService, wRPS, wError := getWeights(appType)
	numNodes := len(nodes)

	// 构建邻接表并计算边权
	adj := make([][]computeEdge, numNodes)
	for i := range trafficMap {
		for j := range trafficMap[i] {
			if trafficMap[i][j].isConnected {
				edge := trafficMap[i][j]
				weight := wService*float64(edge.serviceNum) + wRPS*edge.rps + wError*edge.errorRate
				adj[i] = append(adj[i], computeEdge{j, weight})
			}
		}
	}

	// 拓扑排序
	order := topologicalSort(adj)
	if len(order) != numNodes {
		return 0, nil, nil // 存在环，无法处理
	}

	// 初始化距离和前驱节点
	dist := make([]float64, numNodes)
	prev := make([]int, numNodes)
	for i := range dist {
		dist[i] = -math.MaxFloat64
		prev[i] = -1
	}
	for _, root := range rootNodes {
		if root.num < numNodes {
			dist[root.num] = 0
		}
	}

	// 动态规划处理最长路径
	for _, u := range order {
		if dist[u] == -math.MaxFloat64 {
			continue
		}
		for _, e := range adj[u] {
			if dist[e.to] < dist[u]+e.weight {
				dist[e.to] = dist[u] + e.weight
				prev[e.to] = u
			}
		}
	}

	// 找到最大距离的节点
	maxSum = -math.MaxFloat64
	endNode := -1
	for i, d := range dist {
		if d > maxSum {
			maxSum = d
			endNode = i
		}
	}

	if endNode == -1 || maxSum == -math.MaxFloat64 {
		return 0, nil, nil
	}

	// 回溯路径
	path = []string{}
	for current := endNode; current != -1; current = prev[current] {
		path = append([]string{nodes[current]}, path...)
	}

	n := len(path)
	criticalPathNodeMetrics = make([]CriticalPathNodeMetric, n)
	for i := 0; i+1 < n; i++ {
		strat := path[i]
		end := path[i+1]
		stratNum := -1
		endNum := -1
		for j := 0; j < len(nodes); j++ {
			if nodes[j] == strat {
				stratNum = j
			}
			if nodes[j] == end {
				endNum = j
			}
		}
		criticalPathNodeMetrics[i+1].ServiceNum = trafficMap[stratNum][endNum].serviceNum
		criticalPathNodeMetrics[i+1].Rps = trafficMap[stratNum][endNum].rps
		criticalPathNodeMetrics[i+1].ErrorRate = trafficMap[stratNum][endNum].errorRate
	}
	return maxSum, path, criticalPathNodeMetrics
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
