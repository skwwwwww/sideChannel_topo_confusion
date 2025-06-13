package topo

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Root 结构体
type Root struct {
	Timestamp int      `json:"timestamp"`
	Duration  int      `json:"duration"`
	GraphType string   `json:"graphType"`
	Elements  Elements `json:"elements"`
}

// Elements 结构体
type Elements struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// 目前Url -> Json的比较完整了
func GetTopo() Root {
	//这里我觉得可以改成一个常量？
	// url := "http://kiali.istio-system.svc.cluster.local:20001/kiali/api/namespaces/graph?duration=60s&graphType=versionedApp&includeIdleEdges=true&injectServiceNodes=true&boxBy=cluster,namespace,app&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health,idleNode&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=sockshop-coherence,sockshop-core,default,istio-system,local-path-storage"
	// test_url := "http://192.168.200.153:20001/kiali/api/namespaces/graph?duration=60s&graphType=versionedApp&includeIdleEdges=true&injectServiceNodes=true&boxBy=cluster,namespace,app&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health,idleNode&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=sockshop-coherence,sockshop-core,default,istio-system,local-path-storage"
	// 这个包括了service以及workload
	// test_url1 := "http://192.168.200.153:20001/kiali/api/namespaces/graph?duration=60s&graphType=versionedApp&includeIdleEdges=false&injectServiceNodes=true&boxBy=app&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health&rateGrpc=requests&rateHttp=requests&rateTcp=total&namespaces=istio-system,local-path-storage,sockshop-coherence,sockshop-core,default"
	// 这个只包含了service上的路径
	test_url1 := "http://192.168.88.150:20001/kiali/api/namespaces/graph?duration=60s&graphType=service&includeIdleEdges=false&injectServiceNodes=true&boxBy=cluster,namespace&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health,idleNode&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=default"
	// test_url1 := "http://kiali.istio-system.svc.cluster.local:20001/kiali/api/namespaces/graph?duration=60s&graphType=service&includeIdleEdges=false&injectServiceNodes=true&boxBy=cluster,namespace&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=default,istio-system,local-path-storage,sockshop-coherence,sockshop-core"
	resp, err := http.Get(test_url1)
	if err != nil {
		log.Fatalf("Error fetching data: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应数据
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// 将响应的 JSON 数据解析到结构体
	var root Root
	err = json.Unmarshal(body, &root)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return root
}
