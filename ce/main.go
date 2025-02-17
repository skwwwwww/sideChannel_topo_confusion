// main.go

package main

import (
	// "fmt"

	// "html/template"
	// criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
	generaltg "github.com/sideChannel_topo_confusion/ce/generalTG"
	// obfuscation "github.com/sideChannel_topo_confusion/ce/obfuscationstrategy"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/tools/clientcmd"
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
	select {}

}
