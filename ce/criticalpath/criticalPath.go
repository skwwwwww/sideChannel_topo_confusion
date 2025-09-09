package criticalpath

import (
	"encoding/json"
	"log"
	"time" // 引入 time 包

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

type PathInfo struct {
	Nodes        []string
	SumDegree    int
	KeyNodeCount int
}

func GetCriticalPaths() (float64, []string, map[string]TrafficNode, []CriticalPathNodeMetric) {
	// 1. 获取原始拓扑
	// 记录关键路径计算开始时间

	root := topo.GetTopo()
	jsonData, _ := json.MarshalIndent(root, "", "  ") // 使用两个空格作为缩进
	log.Println("获取到的原始拓扑:", string(jsonData))

	root = FilterTopoByServiceAndTrafficServiceNames(root)

	// 改进点：使用 MarshalIndent
	filteredJsonData, err := json.MarshalIndent(root, "", "  ") // 同样使用两个空格缩进
	if err != nil {
		log.Printf("Failed to marshal filtered topo: %v", err)
	} else {
		log.Println("过滤后的拓扑:", string(filteredJsonData))
	}

	trafficMap, nodes, nodesMap, rootNodes := shaped1(root)
	log.Println("获取到的流量映射:", trafficMap)
	log.Println("获取到的节点:", nodes)
	log.Println("获取到的节点映射:", nodesMap)
	log.Println("获取到的根节点:", rootNodes)

	// 2. 计算关键路径并返回
	startTime := time.Now()
	maxSum, path, criticalPathNodeMetrics := FindCriticalPath(trafficMap, nodes, nodesMap, rootNodes, SERVICE_DEPENDENCY)

	// 记录关键路径计算结束时间
	endTime := time.Now()

	// 计算并输出关键路径计算所消耗的时间
	duration := endTime.Sub(startTime)
	log.Printf("关键路径分析消耗的时间: %v\n", duration)

	return maxSum, path, nodesMap, criticalPathNodeMetrics
}
