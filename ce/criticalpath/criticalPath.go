package criticalpath

import (
	"encoding/json"
	"log"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

type PathInfo struct {
	Nodes        []string
	SumDegree    int
	KeyNodeCount int
}

func GetCriticalPaths() (float64, []string, map[string]TrafficNode, []CriticalPathNodeMetric) {
	// 1. 获取原始拓扑
	root := topo.GetTopo()
	jsonData, _ := json.Marshal(root)
	log.Println("获取到的原始拓扑:", string(jsonData))

	root = FilterTopoByServiceAndTrafficServiceNames(root)

	jsonData, _ = json.Marshal(root)
	log.Println("过滤后的拓扑:", string(jsonData))

	trafficMap, nodes, nodesMap, rootNodes := shaped1(root)
	log.Println("获取到的流量映射:", trafficMap)
	log.Println("获取到的节点:", nodes)
	log.Println("获取到的节点映射:", nodesMap)
	log.Println("获取到的根节点:", rootNodes)

	// 2. 计算关键路径并返回
	maxSum, path, criticalPathNodeMetrics := FindCriticalPath(trafficMap, nodes, nodesMap, rootNodes, SERVICE_DEPENDENCY)
	return maxSum, path, nodesMap, criticalPathNodeMetrics
}
