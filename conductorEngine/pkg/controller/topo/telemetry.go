package topo

func getTrafficMap() graph.TrafficMap {

	return BuildNamespacesTrafficMap(ctx, o.TelemetryOptions, globalInfo)
}
