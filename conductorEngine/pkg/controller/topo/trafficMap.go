package topo

import (
	"github.com/sideChannel_topo_confusion/conductorEngine/pkg/controler/graph"
)

func BuildNamespacesTrafficMap(o graph.TelemetryOptions) {
	trafficMap := graph.NewTrafficMap()

	for _, namespace := range o.Namespaces {
		namespaceTrafficMap := buildNamespaceTrafficMap(ctx, namespace.Name, o, globalInfo)

		// Merge this namespace into the final TrafficMap
		telemetry.MergeTrafficMaps(trafficMap, namespace.Name, namespaceTrafficMap)
	}

}

func buildNamespaceTrafficMap() {

}
