package criticalpath

import (
	"encoding/json"
	"fmt"

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
	s, _ := json.Marshal(root.Elements.Edges)
	fmt.Println(string(s))

	root = ServiceFilter(root)

	trafficMap, nodes, nodesMap, rootNodes := shaped1(root)
	s, _ = json.Marshal(trafficMap)
	fmt.Println(string(s))
	s, _ = json.Marshal(nodes)
	fmt.Println(string(s))
	s, _ = json.Marshal(nodesMap)
	fmt.Println(string(s))
	s, _ = json.Marshal(rootNodes)
	fmt.Println(string(s))

	// 2. 计算关键路径并返回
	maxSum, path, criticalPathNodeMetrics := FindCriticalPath(trafficMap, nodes, nodesMap, rootNodes, SERVICE_DEPENDENCY)
	return maxSum, path, nodesMap, criticalPathNodeMetrics
}
