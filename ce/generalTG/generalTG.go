package generaltg

import (
	"context"
	"fmt"
	"log"

	criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// 返回的是不同节点的下游节点
func GeneralTrafficGenertator() {
	topo, _, keyPaths := criticalpath.GetCriticalPaths()

	namespace := topo.Nodes[keyPaths[0].Nodes[0]].Data.Namespace
	// 加载 kubeconfig 文件（默认路径为 ~/.kube/config）
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Failed to load kubeconfig: %v", err)
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// 定义要创建的实例个数
	instanceCount := len(keyPaths)
	// 为每个实例创建 Deployment 和 Service
	for i := 1; i <= instanceCount; i++ {
		instanceID := string(rune('a' - 1 + i)) // 生成唯一的实例 ID，例如 a, b, c

		// 创建 Deployment
		deployment := createDeployment(instanceID)
		_, err = clientset.AppsV1().Deployments(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Failed to create deployment %s: %v", instanceID, err)
			continue
		}
		fmt.Printf("Deployment created successfully: traffic-generator-%s\n", instanceID)

		// 创建 Service
		service := createService(instanceID)
		_, err = clientset.CoreV1().Services("default").Create(context.TODO(), service, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Failed to create service %s: %v", instanceID, err)
			continue
		}
		fmt.Printf("Service created successfully: traffic-service-%s\n", instanceID)
	}

}

// 创建 Deployment
func createDeployment(instanceID string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-generator-" + instanceID,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "traffic-generator-" + instanceID,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "traffic-generator-" + instanceID,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "traffic-generator",
							Image: "traffic-generator:1.0",
							Env: []corev1.EnvVar{
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}
}

// 创建 Service
func createService(instanceID string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "traffic-service-" + instanceID,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "traffic-generator-" + instanceID,
			},
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// 辅助函数：返回 int32 指针
func int32Ptr(i int32) *int32 {
	return &i
}
