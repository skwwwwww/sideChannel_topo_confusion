package criticalpath

import (
	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

type Topo struct {
	Nodes map[string]topo.Node
	Edges map[string]topo.Edge
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
