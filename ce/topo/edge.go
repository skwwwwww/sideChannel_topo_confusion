package topo

// Edge 结构体
type Edge struct {
	Data EdgeData `json:"data"`
}

// EdgeData 结构体
type EdgeData struct {
	ID      string      `json:"id"`
	Source  string      `json:"source"`
	Target  string      `json:"target"`
	Traffic EdgeTraffic `json:"traffic"`
}

// Traffic 结构体
type EdgeTraffic struct {
	Protocol  string    `json:"protocol"`
	Rates     EdgeRate  `json:"rates"`
	Responses Status200 `json:"200"`
}

type EdgeRate struct {
	Http           string `json:"http"`
	HttpPercentReq string `json:"httpPercentReq"`
}

type Status200 struct {
	Flags map[string]float64 `json:"flags"`
	Hosts map[string]float64 `json:"hosts"`
}
