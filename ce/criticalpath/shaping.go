package criticalpath

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

const (
	SERVICE_DEPENDENCY   = 1
	DELAY_SENSITIVE      = 2
	TRAFFIC_INTENSIVE    = 3
	RELIABILITY_PRIORITY = 4
	RESOUTRCE_INTENSIVE  = 5

	TRAFFIC = "traffic"
	METRICS = "metrics"
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

// 将从kiali获取到的topo重塑成加权有向连接矩阵
func shaped1(root topo.Root) (TrafficMap [][]float64, nodes []string) {

	return [][]float64{}, []string{}
}

// 根据不同的应用类型返回不同的权值
// 服务依赖性 -> 边的权值为 "1"
// 延迟敏感型 -> 边的权值为服务间的平均延迟
// 流量密集型 -> 边的权值为服务间的流量大小（这个不知道kiali中能不能获取到）
// 可靠优先型 -> 边的权值为错误率
// 资源密集型 -> 待定
func GetValue(applicationType int, namespace string, service string) float64 {
	if applicationType == SERVICE_DEPENDENCY {
		return 1.0
	}
	if applicationType == DELAY_SENSITIVE {

	}
	if applicationType == TRAFFIC_INTENSIVE {

	}
	if applicationType == RELIABILITY_PRIORITY {

	}
	return 1.0
}

// metricTpye 目前有两种TRAFFIC和METRICS
func GetTrafficMertics(namespace, service, metricTpye string) {

	url := "http://192.168.200.153:20001/kiali/console/namespaces/" + namespace + "/services/" + service + "?duration=60&refresh=60000&tab=" + metricTpye + "&rangeDuration=600"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if metricTpye == METRICS {
		var metrics InboundMetrics
		if err := json.Unmarshal(body, &metrics); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v", metrics)
	}

	if metricTpye == TRAFFIC {
		var topology ServiceTopology
		if err := json.Unmarshal(body, &topology); err != nil {
			log.Fatal(err)
		}
		// 访问数据示例
		fmt.Println("第一个节点的应用名称:", topology.Elements.Nodes[0].Data.App)
		fmt.Println("第一条边的协议:", topology.Elements.Edges[0].Data.Traffic.Protocol)
	}

}
