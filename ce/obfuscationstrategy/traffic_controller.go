package obfuscation

import (
	"fmt"

	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/clientset/istiov1alpha3"
	"istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 创建混淆流量路由规则
func CreateObfuscationRoute(istioClient versioned.Interface, originalSvc string, obfuscationSvcs []string) error {
	// 定义VirtualService
	vs := &networkingv1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-obfuscation", originalSvc),
		},
		Spec: istiov1alpha3.VirtualService{
			Hosts:    []string{originalSvc},
			Gateways: []string{"mesh"},
			Http: []*istiov1alpha3.HTTPRoute{
				// 真实业务流量路由（95%）
				{
					Route: []*istiov1alpha3.HTTPRouteDestination{{
						Destination: &istiov1alpha3.Destination{
							Host: originalSvc,
						},
						Weight: 95,
					}},
				},
				// 混淆流量路由（5%）
				{
					Match: []*istiov1alpha3.HTTPMatchRequest{{
						Headers: map[string]*istiov1alpha3.StringMatch{
							"X-Obfs-Mark": {MatchType: &istiov1alpha3.StringMatch_Exact{Exact: "true"}},
						},
					}},
					Route: generateObfuscationRoutes(obfuscationSvcs),
				},
			},
		},
	}

	// 应用配置
	_, err := istioClient.NetworkingV1alpha3().VirtualServices("default").Create(vs)
	return err
}

// 生成混淆路由链
func generateObfuscationRoutes(svcs []string) []*istiov1alpha3.HTTPRouteDestination {
	var routes []*istiov1alpha3.HTTPRouteDestination
	weight := 100 / len(svcs) // 平均分配权重

	for _, svc := range svcs {
		routes = append(routes, &istiov1alpha3.HTTPRouteDestination{
			Destination: &istiov1alpha3.Destination{
				Host: svc,
			},
			Weight: int32(weight),
		})
	}
	return routes
}
