package graph

import "time"

type Node struct {
	ID        string  // unique identifier for the node
	NodeType  string  // Node type
	Cluster   string  // Cluster
	Namespace string  // Namespace
	Workload  string  // Workload (deployment) name
	App       string  // Workload app label value
	Version   string  // Workload version label value
	Service   string  // Service name
	Edges     []*Edge // child nodes

	Metadata Metadata // app-specific data general purpose map
	// for holding any desired node or edge information
}

type Edge struct {
	Source   *Node
	Dest     *Node
	Metadata Metadata // app-specific data
}

// TrafficMap is a map of app Nodes, each optionally holding Edge data. Metadata
// is a general purpose map for holding any desired node or edge information.
// Each app node should have a unique namespace+workload.  Note that it is feasible
// but likely unusual to have two nodes with the same name+version in the same
// namespace.
type TrafficMap map[string]*Node

type NamespaceInfo struct {
	Name     string
	Duration time.Duration
	IsIstio  bool
}

type NamespaceInfoMap map[string]NamespaceInfo

// NewTrafficMap constructor
func NewTrafficMap() TrafficMap {
	return make(map[string]*Node)
}
