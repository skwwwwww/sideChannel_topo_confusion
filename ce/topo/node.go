package topo

// Node 结构体
type Node struct {
	Data NodeData `json:"data"`
}

// NodeData 结构体
type NodeData struct {
	//通用
	App        string     `json:"app,omitempty"`
	Cluster    string     `json:"cluster"`
	HealthData HealthData `json:"healthData,omitempty"`
	ID         string     `json:"id"`
	Namespace  string     `json:"namespace"`
	NodeType   string     `json:"nodeType"` //三种box, app, service

	//service app
	DestServices []DestService `json:"destServices,omitempty"`
	Parent       string        `json:"parent"`
	Traffic      []NodeTraffic `json:"traffic,omitempty"`

	//service
	Service string `json:"service,omitempty"`

	//app
	Version  string `json:"version,omitempty"`
	Workload string `json:"workload,omitempty"`

	//box
	IsBox string `json:"isBox"` //service  没有这个字段

	//只有根节点有（很有用滴）
	IsRoot bool `json:"isRoot,omitempty"`
}

// HealthData 结构体
type HealthData struct {
	//只有box有，里面有各个workload的Replicas的相关数据
	WorkloadStatuses []WorkloadStatus `json:"workloadStatuses,omitempty"`

	Requests Requests `json:"requests"`
}

// WorkloadStatus 结构体
type WorkloadStatus struct {
	Name              string `json:"name"`
	DesiredReplicas   int    `json:"desiredReplicas"`
	CurrentReplicas   int    `json:"currentReplicas"`
	AvailableReplicas int    `json:"availableReplicas"`
	SyncedProxies     int    `json:"syncedProxies"`
}

// Requests 结构体(这里有问题，需要之后仔细考量一下嘿嘿嘿)
type Requests struct {
	Inbound           Inbound                `json:"inbound,omitempty"`
	Outbound          Outbound               `json:"outbound,omitempty"`
	HealthAnnotations map[string]interface{} `json:"healthAnnotations"`
}

type Inbound struct {
	Http Http `json:"http"`
}

type Outbound struct {
	Http Http `json:"http"`
}

type Http struct {
	HttpRequestsPerSecond200 float64 `json:"200"`
}

// DestService 结构体
type DestService struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type NodeTraffic struct {
	Protocol string   `json:"protocol"`
	Rates    NodeRate `json:"rates"`
}

type NodeRate struct {
	HttpIn  string `json:"httpIn,omitempty"`
	HttpOut string `json:"httpOut,omitempty"`
}
