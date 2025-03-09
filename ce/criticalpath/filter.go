package criticalpath

import (
	"strings"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

// 将用于混淆的servcie以及过滤掉
func ServiceFilter(t topo.Root) topo.Root {
	nodeMap := map[string]bool{}
	toRemoveNode := []int{}
	toRemoveEdge := []int{}

	for i, v := range t.Elements.Nodes {
		if strings.HasPrefix(v.Data.Service, "traffic-generator") || strings.HasPrefix(v.Data.Workload, "traffic-generator") {
			nodeMap[v.Data.Service] = true
			toRemoveNode = append(toRemoveNode, i)
			//t.Elements.Nodes = append(t.Elements.Nodes[:i], t.Elements.Nodes[i+1:]...)
		}
	}
	// 从后往前删除元素
	for i := len(toRemoveNode) - 1; i >= 0; i-- {
		t.Elements.Nodes = append(t.Elements.Nodes[:toRemoveNode[i]], t.Elements.Nodes[toRemoveNode[i]+1:]...)
	}
	for i, v := range t.Elements.Edges {
		if nodeMap[v.Data.Source] || nodeMap[v.Data.Target] {
			toRemoveEdge = append(toRemoveEdge, i)
			//t.Elements.Edges = append(t.Elements.Edges[:i], t.Elements.Edges[i+1:]...)
		}
	}
	// 从后往前删除元素
	for i := len(toRemoveEdge) - 1; i >= 0; i-- {
		t.Elements.Edges = append(t.Elements.Edges[:toRemoveEdge[i]], t.Elements.Edges[toRemoveEdge[i]+1:]...)
	}
	return t
}
