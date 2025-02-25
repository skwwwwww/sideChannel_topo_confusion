package criticalpath

type InboundMetrics struct {
	Name          string        `json:"name"`
	Title         string        `json:"title"`
	Charts        []Chart       `json:"charts"`
	Aggregations  []Aggregation `json:"aggregations"`
	ExternalLinks interface{}   `json:"externalLinks"`
	Rows          int           `json:"rows"`
}

type Chart struct {
	Name           string      `json:"name"`
	Unit           string      `json:"unit"`
	Spans          int         `json:"spans"`
	StartCollapsed bool        `json:"startCollapsed"`
	Metrics        []Metric    `json:"metrics"`
	XAxis          interface{} `json:"xAxis"`
	Error          string      `json:"error"`
}

type Metric struct {
	Labels     map[string]string `json:"labels"`
	Datapoints [][]interface{}   `json:"datapoints"`
	Stat       *string           `json:"stat,omitempty"`
	Name       string            `json:"name"`
}

type Aggregation struct {
	Label           string `json:"label"`
	DisplayName     string `json:"displayName"`
	SingleSelection bool   `json:"singleSelection"`
}

// ////////////////////////////////////////////////////
type ServiceTopology struct {
	Timestamp int64    `json:"timestamp"`
	Duration  int      `json:"duration"`
	GraphType string   `json:"graphType"`
	Elements  Elements `json:"elements"`
}

type Elements struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Node struct {
	Data NodeData `json:"data"`
}

type NodeData struct {
	ID           string        `json:"id"`
	NodeType     string        `json:"nodeType"`
	Cluster      string        `json:"cluster"`
	Namespace    string        `json:"namespace"`
	App          string        `json:"app"`
	Service      *string       `json:"service,omitempty"`
	Workload     *string       `json:"workload,omitempty"`
	Version      *string       `json:"version,omitempty"`
	DestServices []DestService `json:"destServices"`
	Traffic      []Traffic     `json:"traffic"`
	HealthData   interface{}   `json:"healthData"`
	IsRoot       *bool         `json:"isRoot,omitempty"`
}

type DestService struct {
	Cluster   string `json:"cluster"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type Traffic struct {
	Protocol string `json:"protocol"`
	Rates    Rates  `json:"rates"`
}

type Rates struct {
	HttpIn         *string `json:"httpIn,omitempty"`
	HttpOut        *string `json:"httpOut,omitempty"`
	Http           *string `json:"http,omitempty"`
	HttpPercentReq *string `json:"httpPercentReq,omitempty"`
}

type Edge struct {
	Data EdgeData `json:"data"`
}

type EdgeData struct {
	ID              string      `json:"id"`
	Source          string      `json:"source"`
	Target          string      `json:"target"`
	DestPrincipal   string      `json:"destPrincipal"`
	IsMTLS          string      `json:"isMTLS"`
	SourcePrincipal string      `json:"sourcePrincipal"`
	ResponseTime    *string     `json:"responseTime,omitempty"`
	Throughput      *string     `json:"throughput,omitempty"`
	Traffic         EdgeTraffic `json:"traffic"`
}

type EdgeTraffic struct {
	Protocol  string                  `json:"protocol"`
	Rates     Rates                   `json:"rates"`
	Responses map[string]ResponseCode `json:"responses"`
}

type ResponseCode struct {
	Flags map[string]string `json:"flags"`
	Hosts map[string]string `json:"hosts"`
}
