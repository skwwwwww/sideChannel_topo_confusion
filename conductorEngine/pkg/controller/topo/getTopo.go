package topo

import (
	"context"
	"crypto/md5"
	"fmt"
	"regexp"
	"time"

	prom_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type TrafficMap map[string]*Node

type Node struct {
	ID       string
	Metadata map[string]interface{}
	Edges    []*Edge
}

type Edge struct {
	Dest     *Node
	Metadata map[string]interface{}
}

const (
	ProtocolKey = "protocol"
)

var grpcMetric = regexp.MustCompile(`istio_.*_messages`)

// AddEdge creates an edge from a source to a destination node
func (n *Node) AddEdge(dest *Node) *Edge {
	edge := &Edge{
		Dest:     dest,
		Metadata: make(map[string]interface{}),
	}
	n.Edges = append(n.Edges, edge)
	return edge
}

// GenerateTrafficMap builds a traffic map by querying Prometheus metrics
func GenerateTrafficMap(promApi prom_v1.API, namespace string, queryTime time.Time, duration time.Duration) TrafficMap {
	trafficMap := make(TrafficMap)

	// Define query
	metric := "istio_requests_total"
	groupBy := "source_workload_namespace,source_workload,destination_workload_namespace,destination_workload,request_protocol,response_code"
	query := fmt.Sprintf(`sum(rate(%s{destination_workload_namespace="%s"}[%vs])) by (%s)`,
		metric,
		namespace,
		int(duration.Seconds()*10),
		groupBy,
	)

	// Execute Prometheus query
	vector := promQuery(query, queryTime, promApi)
	populateTrafficMap(trafficMap, &vector, metric)

	return trafficMap
}

// promQuery executes a Prometheus query and returns the result vector
func promQuery(query string, queryTime time.Time, promApi prom_v1.API) model.Vector {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("Executing query: %s\n", query)
	value, warnings, err := promApi.Query(ctx, query, queryTime)
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		return nil
	}
	if len(warnings) > 0 {
		fmt.Printf("Prometheus warnings: %v\n", warnings)
	}

	if value.Type() == model.ValVector {
		return value.(model.Vector)
	}
	return nil
}

// populateTrafficMap processes Prometheus vector data and populates the traffic map
func populateTrafficMap(trafficMap TrafficMap, vector *model.Vector, metric string) {
	for _, sample := range *vector {
		labels := sample.Metric
		val := float64(sample.Value)

		// Extract labels
		sourceNamespace := string(labels["source_workload_namespace"])
		sourceWorkload := string(labels["source_workload"])
		destNamespace := string(labels["destination_workload_namespace"])
		destWorkload := string(labels["destination_workload"])
		protocol := string(labels["request_protocol"])
		responseCode := string(labels["response_code"])

		// Generate node IDs
		sourceID := fmt.Sprintf("%s.%s", sourceNamespace, sourceWorkload)
		destID := fmt.Sprintf("%s.%s", destNamespace, destWorkload)

		// Add or get source node
		sourceNode, exists := trafficMap[sourceID]
		if !exists {
			sourceNode = &Node{
				ID:       sourceID,
				Metadata: make(map[string]interface{}),
			}
			trafficMap[sourceID] = sourceNode
		}

		// Add or get destination node
		destNode, exists := trafficMap[destID]
		if !exists {
			destNode = &Node{
				ID:       destID,
				Metadata: make(map[string]interface{}),
			}
			trafficMap[destID] = destNode
		}

		// Add edge between source and destination
		addEdgeTraffic(val, protocol, responseCode, sourceNode, destNode, metric)
	}
}

// addEdgeTraffic adds traffic data to an edge between two nodes
func addEdgeTraffic(val float64, protocol, responseCode string, source, dest *Node, metric string) {
	var edge *Edge
	for _, e := range source.Edges {
		if e.Dest.ID == dest.ID && e.Metadata[ProtocolKey] == protocol {
			edge = e
			break
		}
	}
	if edge == nil {
		edge = source.AddEdge(dest)
		edge.Metadata[ProtocolKey] = protocol
	}

	// Add traffic data to edge
	hash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s", metric, source.ID, dest.ID, responseCode))))
	if _, exists := edge.Metadata[hash]; !exists {
		edge.Metadata[hash] = true
		edge.Metadata["traffic"] = val
	}
}
