package criticalpath

import (
	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

type PathInfo struct {
	Nodes        []string
	SumDegree    int
	KeyNodeCount int
}

// 1. 计算节点的度-> 改进：将不同的指标赋予不同的权值，求和后赋值给边
func calculateDegrees(t Topo) map[string]int {
	degrees := make(map[string]int)
	for _, edge := range t.Edges {
		degrees[edge.Data.Source]++
		degrees[edge.Data.Target]++
	}
	return degrees
}

// 2. 找到图中度最大的节点
func findKeyNodes(degrees map[string]int) map[string]bool {
	maxDegree := 0
	for _, d := range degrees {
		if d > maxDegree {
			maxDegree = d
		}
	}

	keyNodes := map[string]bool{}
	for node, d := range degrees {
		if d == maxDegree {
			keyNodes[node] = true
		} else {
			keyNodes[node] = false
		}
	}

	return keyNodes
}

// 3. 枚举所有边
func generateAllPaths(topo Topo, degrees map[string]int, keyNodes map[string]bool) []PathInfo {
	var paths []PathInfo
	for nodeID := range topo.Nodes {
		visited := make(map[string]bool)
		visited[nodeID] = true
		findAllPaths(topo,
			[]string{nodeID},
			degrees[nodeID],
			boolToInt(keyNodes[nodeID]),
			visited,
			&paths,
			degrees,
			keyNodes)
	}
	return paths
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func findAllPaths(topo Topo, currentPath []string, currentSum int, currentKeyCount int, visited map[string]bool, paths *[]PathInfo, degrees map[string]int, keyNodes map[string]bool) {
	currentNode := currentPath[len(currentPath)-1]

	// 记录有效路径（至少包含两个节点）
	if len(currentPath) >= 2 {
		*paths = append(*paths, PathInfo{
			Nodes:        append([]string{}, currentPath...),
			SumDegree:    currentSum,
			KeyNodeCount: currentKeyCount,
		})
	}

	// 遍历出边
	for _, edge := range topo.Edges {
		if edge.Data.Source == currentNode && !visited[edge.Data.Target] {
			newVisited := make(map[string]bool)
			for k, v := range visited {
				newVisited[k] = v
			}
			newVisited[edge.Data.Target] = true

			newPath := append(currentPath, edge.Data.Target)
			newSum := currentSum + degrees[edge.Data.Target]
			newKeyCount := currentKeyCount
			if keyNodes[edge.Data.Target] {
				newKeyCount++
			}

			findAllPaths(topo, newPath, newSum, newKeyCount, newVisited, paths, degrees, keyNodes)
		}
	}
}

// 寻找关键路径
func findKeyPaths(allPaths []PathInfo) []PathInfo {
	if len(allPaths) == 0 {
		return nil
	}

	// 找最大度数总和
	maxSum := 0
	for _, path := range allPaths {
		if path.SumDegree > maxSum {
			maxSum = path.SumDegree
		}
	}

	// 筛选候选路径
	var candidates []PathInfo
	for _, path := range allPaths {
		if path.SumDegree == maxSum {
			candidates = append(candidates, path)
		}
	}

	// 找最大关键节点数
	maxKeyCount := 0
	for _, path := range candidates {
		if path.KeyNodeCount > maxKeyCount {
			maxKeyCount = path.KeyNodeCount
		}
	}

	// 最终筛选
	var keyPaths []PathInfo
	for _, path := range candidates {
		if path.KeyNodeCount == maxKeyCount {
			keyPaths = append(keyPaths, path)
		}
	}
	return keyPaths
}

func GetCriticalPaths() (Topo, []string, []PathInfo) {
	// 1. 获取原始拓扑
	root := topo.GetTopo()
	topo := Shaped(root)

	// 2. 计算关键节点和路径
	degrees := calculateDegrees(topo)
	keyNodes := findKeyNodes(degrees)
	keyPaths := findKeyPaths(generateAllPaths(topo, degrees, keyNodes))
	criticalNodes := []string{}
	for k, v := range keyNodes {
		if v == true {
			criticalNodes = append(criticalNodes, k)
		}
	}
	return topo, criticalNodes, keyPaths
}
