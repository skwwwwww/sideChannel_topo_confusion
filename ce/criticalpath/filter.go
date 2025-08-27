package criticalpath

import (
	"log"
	"strings"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

func FilterTopoByServiceAndTrafficServiceNames(rawRoot topo.Root) topo.Root {
	filteredRoot := topo.Root{
		Timestamp: rawRoot.Timestamp,
		Duration:  rawRoot.Duration,
		GraphType: rawRoot.GraphType,
		Elements:  topo.Elements{}, // 初始化 Elements
	}
	// 定义需要过滤的关键字和前缀（不区分大小写）
	trafficServicePrefix := "traffic-service"
	// `ce` 和 `passthroughcluster` 需要在 ID, App, Service, Workload 字段中进行包含匹配
	blacklistKeywords := []string{"ce", "passthroughcluster"}

	// Step 1: 过滤节点，并记录保留的节点的 ID
	var keptNodes []topo.Node
	keptNodeIDs := make(map[string]struct{}) // 使用 map 来快速查找节点 ID

	for _, node := range rawRoot.Elements.Nodes {
		nodeType := node.Data.NodeType
		nodeID := node.Data.ID
		nodeApp := node.Data.App
		nodeService := node.Data.Service
		nodeWorkload := node.Data.Workload

		// 1.1 检查节点类型是否为 "service" 或 "workload"
		isAllowedType := (nodeType == "service" || nodeType == "workload")
		if !isAllowedType {
			log.Printf("过滤掉节点 (非 service 或 workload 类型): ID: %s, Type: %s, App: %s, Service: %s, Workload: %s\n",
				nodeID, nodeType, nodeApp, nodeService, nodeWorkload)
			continue // 跳过当前节点，不添加到保留列表
		}

		// 1.2 检查是否符合特定黑名单规则
		shouldFilter := false
		filterReason := ""

		// Rule A: 过滤掉名称前缀带有 "traffic-service" 的节点
		// 检查 Service, App, Workload, ID 字段
		if strings.HasPrefix(nodeService, trafficServicePrefix) {
			shouldFilter = true
			filterReason = "Service 字段以 '" + trafficServicePrefix + "' 开头"
		} else if strings.HasPrefix(nodeApp, trafficServicePrefix) {
			shouldFilter = true
			filterReason = "App 字段以 '" + trafficServicePrefix + "' 开头"
		} else if strings.HasPrefix(nodeWorkload, trafficServicePrefix) {
			shouldFilter = true
			filterReason = "Workload 字段以 '" + trafficServicePrefix + "' 开头"
		}

		// Rule B: 检查是否包含黑名单关键字（"ce", "PassthroughCluster"）
		// 只有在 Rule A 没有触发过滤时才继续检查 Rule B
		if !shouldFilter {
			// 将相关字段转换为小写以便进行不区分大小写的包含检查
			lowerNodeApp := strings.ToLower(nodeApp)
			lowerNodeService := strings.ToLower(nodeService)
			lowerNodeWorkload := strings.ToLower(nodeWorkload)

			for _, keyword := range blacklistKeywords {
				lowerKeyword := strings.ToLower(keyword) // 确保关键词也是小写
				if strings.Contains(lowerNodeApp, lowerKeyword) ||
					strings.Contains(lowerNodeService, lowerKeyword) ||
					strings.Contains(lowerNodeWorkload, lowerKeyword) {
					shouldFilter = true
					filterReason = "包含黑名单关键字: '" + keyword + "'"
					break // 找到一个匹配的关键字就足够
				}
			}
		}

		// 根据过滤结果决定是否保留节点
		if shouldFilter {
			log.Printf("过滤掉节点 (原因: %s): ID: %s, Type: %s, App: %s, Service: %s, Workload: %s\n",
				filterReason, nodeID, nodeType, nodeApp, nodeService, nodeWorkload)
		} else {
			keptNodes = append(keptNodes, node)
			keptNodeIDs[nodeID] = struct{}{} // 记录节点 ID
		}
	}
	filteredRoot.Elements.Nodes = keptNodes
	log.Printf("过滤后保留的节点数量: %d\n", len(keptNodes))

	// Step 2: 过滤边，只保留源和目标都在保留节点列表中的边
	var keptEdges []topo.Edge
	for _, edge := range rawRoot.Elements.Edges {
		edgeID := edge.Data.ID
		edgeSource := edge.Data.Source
		edgeTarget := edge.Data.Target

		// 检查边的源和目标是否都在保留的节点 ID 列表中
		_, okSource := keptNodeIDs[edgeSource]
		_, okTarget := keptNodeIDs[edgeTarget]

		if okSource && okTarget {
			keptEdges = append(keptEdges, edge)
		} else {
			// 详细记录过滤边的原因
			sourceStatus := "不在保留列表"
			if okSource {
				sourceStatus = "在保留列表"
			}
			targetStatus := "不在保留列表"
			if okTarget {
				targetStatus = "在保留列表"
			}
			log.Printf("过滤掉边 (源或目标节点不在保留列表中): Edge ID: %s, Source: %s (%s), Target: %s (%s)\n",
				edgeID, edgeSource, sourceStatus, edgeTarget, targetStatus)
		}
	}
	filteredRoot.Elements.Edges = keptEdges
	log.Printf("过滤后保留的边数量: %d\n", len(keptEdges))

	return filteredRoot
}

// 将用于混淆的servcie以及过滤掉
// func ServiceFilter(t topo.Root) topo.Root {
// 	nodeMap := map[string]bool{}
// 	toRemoveNode := []int{}
// 	toRemoveEdge := []int{}

// 	for i, v := range t.Elements.Nodes {
// 		if strings.HasPrefix(v.Data.Service, "traffic-service") || strings.HasPrefix(v.Data.Workload, "traffic-service") || strings.HasPrefix(v.Data.Service, "ce") || strings.HasPrefix(v.Data.Workload, "ce") || strings.HasPrefix(v.Data.Service, "PassthroughCluster") || strings.HasPrefix(v.Data.Workload, "PassthroughCluster") || strings.HasPrefix(v.Data.Service, "unknown") || strings.HasPrefix(v.Data.Workload, "unknown") {
// 			nodeMap[v.Data.Service] = true
// 			toRemoveNode = append(toRemoveNode, i)
// 			log.Println("过滤掉的节点（Service）:", v.Data.Service)
// 			log.Println("过滤掉的节点（Workload）:", v.Data.Workload)

// 			//t.Elements.Nodes = append(t.Elements.Nodes[:i], t.Elements.Nodes[i+1:]...)
// 		}
// 	}
// 	// 从后往前删除元素
// 	for i := len(toRemoveNode) - 1; i >= 0; i-- {
// 		t.Elements.Nodes = append(t.Elements.Nodes[:toRemoveNode[i]], t.Elements.Nodes[toRemoveNode[i]+1:]...)
// 	}
// 	for i, v := range t.Elements.Edges {
// 		if nodeMap[v.Data.Source] || nodeMap[v.Data.Target] {
// 			toRemoveEdge = append(toRemoveEdge, i)
// 			//t.Elements.Edges = append(t.Elements.Edges[:i], t.Elements.Edges[i+1:]...)
// 		}
// 	}
// 	// 从后往前删除元素
// 	for i := len(toRemoveEdge) - 1; i >= 0; i-- {
// 		t.Elements.Edges = append(t.Elements.Edges[:toRemoveEdge[i]], t.Elements.Edges[toRemoveEdge[i]+1:]...)
// 	}
// 	return t
// }
