// main.go

package main

import (
	// "fmt"

	// "html/template"
	"encoding/json"
	"fmt"
	//"fmt"
	"net/http"

	generaltg "github.com/sideChannel_topo_confusion/ce/generalTG"
)

func main() {
	// 初始化K8s客户端
	// config, _ := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	// k8sClient, _ := kubernetes.NewForConfig(config)

	// 获取原始关键路径
	// topo, criticalNodes, criticalPaths := criticalpath.GetCriticalPaths()

	//获取层数
	//layerCount := obfuscation.CalculateLayerCount(criticalPaths, &topo)

	// 来思考一下，是否要先生成策略，然后交给下一个组件来实现
	// 我选择
	// 部署真实混淆服务
	// for i := 0; i < layerCount; i++ {
	// 	conf := obfuscation.RealServiceConfig{
	// 		Name:       fmt.Sprintf("obfs-layer-%d", i+1),
	// 		Replicas:   2,
	// 		Image:      "obfuscation-proxy:1.2",
	// 		Port:       8080,
	// 		IsHoneypot: i >= layerCount-2, // 最后两层为蜜罐
	// 		Layer:      i + 1,
	// 	}
	// 	obfuscation.DeployRealService(k8sClient, conf)
	// }
	// 生成Deployment配置
	// 	deployTpl := `
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: {{.Name}}-deploy
	// spec:
	//   replicas: {{.Replicas}}
	//   selector:
	//     matchLabels:
	//       app: {{.Name}}
	//   template:
	//     metadata:
	//       labels:
	//         app: {{.Name}}
	//         layer: "{{.Layer}}"
	//         {{- if .IsHoneypot }}
	//         honeypot: "true"
	//         {{- end }}
	//     spec:
	//       containers:
	//       - name: {{.Name}}
	//         image: {{.Image}}
	//         ports:
	//         - containerPort: {{.Port}}
	//         resources:
	//           limits:
	//             cpu: "100m"
	//             memory: "128Mi"
	// ---
	// apiVersion: v1
	// kind: Service
	// metadata:
	//   name: {{.Name}}-svc
	// spec:
	//   selector:
	//     app: {{.Name}}
	//   ports:
	//   - protocol: TCP
	//     port: {{.Port}}
	//     targetPort: {{.Port}}
	// `
	// 	tmpl := template.Must(template.New("deploy").Parse(deployTpl))

	// 创建混淆路由规则
	// originalSvc := criticalPaths[0].Nodes[0]
	// var obfsSvcs []string
	// for i := 0; i < layerCount; i++ {
	// 	obfsSvcs = append(obfsSvcs, fmt.Sprintf("obfs-layer-%d-svc", i+1))
	// }
	// obfuscation.CreateObfuscationRoute(istioClient, originalSvc, obfsSvcs)

	// 构建动态链路
	// links := obfuscation.BuildRealLinks(originalTopo, criticalNodes, criticalPaths)
	// obfuscation.ApplyLinkRules(istioClient, links)
	// _, _, keyPaths := criticalpath.GetCriticalPaths()
	// fmt.Println(keyPaths)
	//len := len(keyPaths)

	generaltg.GeneralTrafficGenertator()

	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/ready", ready)

	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
	for {
	}

	//criticalpath.GetTrafficMertics("default", "ratings", criticalpath.TRAFFIC)
}

// HealthStatus represents the health status of the application.
type HealthStatus struct {
	Status string `json:"status"`
}

// healthz handles the health check request.
func healthz(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查应用程序的健康状态
	if isHealthy() {
		// 如果健康，返回200 OK
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthStatus{Status: "healthy"})
	} else {
		// 如果不健康，返回503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthStatus{Status: "unhealthy"})
	}
}

// ReadinessStatus represents the readiness status of the application.
type ReadinessStatus struct {
	Status string `json:"status"`
}

// ready handles the readiness check request.
func ready(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 检查应用程序的就绪状态
	if isReady() {
		// 如果就绪，返回200 OK
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ReadinessStatus{Status: "ready"})
	} else {
		// 如果未就绪，返回503 Service Unavailable
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ReadinessStatus{Status: "not_ready"})
	}
}

// isReady checks if the application is ready.
// Replace this with your actual readiness check logic.
func isReady() bool {
	// 示例：检查缓存是否已加载
	// return cache.IsLoaded()

	// 这里只是一个示例，返回true表示就绪
	return true
}

// isHealthy checks if the application is healthy.
// Replace this with your actual health check logic.
func isHealthy() bool {
	// 示例：检查数据库连接
	// return db.Ping() == nil

	// 这里只是一个示例，返回true表示健康
	return true
}
