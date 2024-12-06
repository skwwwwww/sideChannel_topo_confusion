package graph

import (
	"net/url"
	"time"
)

// ClusterSensitiveKey is the recommended [string] type for maps keying on a cluster-sensitive name
type ClusterSensitiveKey = string

type AccessibleNamespace struct {
	Cluster           string
	CreationTimestamp time.Time
	Name              string
}

// AccessibleNamepaces is a map with Key: ClusterSensitive namespace Key, Value: *AccessibleNamespace
type AccessibleNamespaces map[ClusterSensitiveKey]*AccessibleNamespace

type RequestedAppenders struct {
	All           bool
	AppenderNames []string
}

type RequestedRates struct {
	Ambient string
	Grpc    string
	Http    string
	Tcp     string
}

type CommonOptions struct {
	Duration  time.Duration
	GraphType string
	Params    url.Values // make available the raw query params for vendor-specific handling
	QueryTime int64      // unix time in seconds
}

type NodeOptions struct {
	Aggregate      string
	AggregateValue string
	App            string
	Cluster        string
	Namespace      string
	Service        string
	Version        string
	Workload       string
}

// TelemetryOptions 是提供给遥测供应商的选项。（这里的供应商感觉不是很准确）
type TelemetryOptions struct {
	AccessibleNamespaces AccessibleNamespaces
	Appenders            RequestedAppenders // requested appenders, nil if param not supplied
	IncludeIdleEdges     bool               // include edges with request rates of 0
	InjectServiceNodes   bool               // inject destination service nodes between source and destination nodes.
	Namespaces           NamespaceInfoMap
	Rates                RequestedRates
	CommonOptions
	NodeOptions
}
