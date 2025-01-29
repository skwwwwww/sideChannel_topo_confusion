package criticalpath

import (
	"fmt"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

// 计算节点的度中心性
func calculateDegreeCentrality(edges map[string]topo.Edge) map[string]int {
	degrees := make(map[string]int)
	for _, edge := range edges {
		degrees[edge.Data.Source]++
		degrees[edge.Data.Target]++
	}
	return degrees
}

// 识别关键节点
func findKeyNodes(degrees map[string]int) []string {
	maxDegree := -1
	for _, degree := range degrees {
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	var keyNodes []string
	for nodeID, degree := range degrees {
		if degree == maxDegree {
			keyNodes = append(keyNodes, nodeID)
		}
	}
	return keyNodes
}

type CriticalPath struct {
	Edges map[string]topo.Edge
	Nodes map[string]topo.Node
}

// 识别关键路径(其实应该加一个参数，用于区别不同的应用类型)
func findCriticalPaths(topology Topo, degrees map[string]int, keyNodes []string) []CriticalPath {
	// 使用DFS来遍历所有可能的路径
	var bestPaths []CriticalPath
	var bestDegreeSum int

	// 为了方便查找关键节点，使用一个集合（map）
	keyNodeSet := make(map[string]struct{})
	for _, nodeID := range keyNodes {
		keyNodeSet[nodeID] = struct{}{}
	}

	// 辅助函数：DFS遍历路径
	var dfs func(nodeID string, visited map[string]bool, path []string, degreeSum int)
	dfs = func(nodeID string, visited map[string]bool, path []string, degreeSum int) {
		// 如果当前节点已被访问过，则返回
		if visited[nodeID] {
			return
		}

		// 标记当前节点为已访问
		visited[nodeID] = true

		// 将当前节点添加到路径中，并累加其度数
		path = append(path, nodeID)
		degreeSum += degrees[nodeID]

		// 如果当前路径包含关键节点，则增加关键节点计数
		keyNodeCount := 0
		for _, node := range path {
			if _, exists := keyNodeSet[node]; exists {
				keyNodeCount++
			}
		}

		if len(bestPaths) == 0 || keyNodeCount > len(bestPaths[0].Nodes) || (keyNodeCount == len(bestPaths[0].Nodes) && degreeSum > bestDegreeSum) {
			// 找到更好的路径，更新最佳路径列表
			if len(bestPaths) == 0 || keyNodeCount > len(bestPaths[0].Nodes) || degreeSum > bestDegreeSum {
				bestPaths = []CriticalPath{}
			}
			// 将新的路径添加到结果中
			criticalPath := CriticalPath{
				Nodes: make(map[string]topo.Node),
				Edges: make(map[string]topo.Edge),
			}
			for i := 0; i < len(path)-1; i++ {
				nodeID := path[i]
				criticalPath.Nodes[nodeID] = topology.Nodes[nodeID]
				// 找到相应的边
				for _, edge := range topology.Edges {
					if edge.Data.Source == nodeID && edge.Data.Target == path[i+1] {
						criticalPath.Edges[edge.Data.ID] = edge
					} else if edge.Data.Target == nodeID && edge.Data.Source == path[i+1] {
						criticalPath.Edges[edge.Data.ID] = edge
					}
				}
			}
			// 添加路径到结果中
			bestPaths = append(bestPaths, criticalPath)
			bestDegreeSum = degreeSum
		}

		// 递归遍历所有相邻的节点
		for _, edge := range topology.Edges {
			if edge.Data.Source == nodeID && !visited[edge.Data.Target] {
				dfs(edge.Data.Target, visited, path, degreeSum)
			} else if edge.Data.Target == nodeID && !visited[edge.Data.Source] {
				dfs(edge.Data.Source, visited, path, degreeSum)
			}
		}

		// 回溯，标记当前节点为未访问
		visited[nodeID] = false
	}

	// 遍历所有节点，尝试从每个节点开始搜索
	for nodeID := range topology.Nodes {
		visited := make(map[string]bool)
		dfs(nodeID, visited, nil, 0)
	}

	// 返回所有找到的关键路径
	return bestPaths
}

func main() {
	root := topo.GetTopo()
	//fmt.Print(root)
	topo := shaping(root)
	degrees := calculateDegreeCentrality(topo.Edges)
	keyNodes := findKeyNodes(degrees)
	//criticalPath := findCriticalPaths(topo, degrees, keyNodes)
	findCriticalPaths(topo, degrees, keyNodes)
	//printCriticalPaths(criticalPath, topo)
}

// Print the best paths
func printCriticalPaths(bestPaths []CriticalPath, topo Topo) {
	for i, path := range bestPaths {
		fmt.Printf("Critical Path %d:\n", i+1)
		fmt.Println("Nodes:")
		for nodeID := range path.Nodes {
			fmt.Println(topo.Nodes[nodeID].Data.Workload)
		}
		fmt.Println("Edges:")
		for _, edge := range path.Edges {
			fmt.Printf("Edge ID: %s, Source: %s, Target: %s\n", edge.Data.ID, topo.Nodes[edge.Data.Source].Data.
				Workload, topo.Nodes[edge.Data.Target].Data.Workload)
		}
		fmt.Println("--------------------------------------------------")
	}
}
