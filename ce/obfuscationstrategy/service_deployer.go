package obfuscation

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

// 真实混淆服务部署配置
type RealServiceConfig struct {
	Name       string
	Replicas   int32
	Image      string
	Port       int32
	Labels     map[string]string
	IsHoneypot bool // 是否为蜜罐服务
	Layer      int  // 所在层级
}

// 部署真实服务
func DeployRealService(client *kubernetes.Clientset, config RealServiceConfig) error {
	// 生成Deployment配置
	deployTpl := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}-deploy
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
        layer: "{{.Layer}}"
        {{- if .IsHoneypot }}
        honeypot: "true"
        {{- end }}
    spec:
      containers:
      - name: {{.Name}}
        image: {{.Image}}
        ports:
        - containerPort: {{.Port}}
        resources:
          limits:
            cpu: "100m"
            memory: "128Mi"
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-svc
spec:
  selector:
    app: {{.Name}}
  ports:
  - protocol: TCP
    port: {{.Port}}
    targetPort: {{.Port}}
`

	// 渲染模板
	tmpl := template.Must(template.New("deploy").Parse(deployTpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return fmt.Errorf("模板渲染失败: %v", err)
	}

	// 分割YAML文档并应用
	decoder := yaml.NewYAMLOrJSONDecoder(&buf, 4096)
	for {
		// 解析Deployment
		var deploy appsv1.Deployment
		var ctx context.Context
		var opts metav1.CreateOptions
		if err := decoder.Decode(&deploy); err != nil {
			break
		}
		if _, err := client.AppsV1().Deployments("default").Create(ctx, &deploy, opts); err != nil {
			return fmt.Errorf("部署Deployment失败: %v", err)
		}

		// 解析Service
		var svc corev1.Service
		if err := decoder.Decode(&svc); err != nil {
			break
		}
		if _, err := client.CoreV1().Services("default").Create(&svc); err != nil {
			return fmt.Errorf("部署Service失败: %v", err)
		}
	}
	return nil
}
