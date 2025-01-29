package obfusionstrategy

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"text/template"
	"time"

	criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
	topo "github.com/sideChannel_topo_confusion/ce/topo"
)

// 混淆配置结构体
type ObfuscationConfig struct {
	VirtualServices []VirtualServiceConfig
	TrafficRules    []TrafficRule
	DecoyPaths      []DecoyPath
}

type VirtualServiceConfig struct {
	Name        string
	TargetHost  string
	MirrorRatio float32
	FakePaths   []string
}

type TrafficRule struct {
	HeaderKey   string
	HeaderValue string
	Probability float32
}

type DecoyPath struct {
	Path          []string
	HoneypotCount int
}

type PathInfo struct {
	Nodes        []string
	SumDegree    int
	KeyNodeCount int
}

// 生成混淆策略的核心函数
func GenerateObfuscation(topo criticalpath.Topo, keyNodes []string, keyPaths []PathInfo) ObfuscationConfig {
	config := ObfuscationConfig{}

	// 根据关键节点生成虚拟服务
	for _, node := range keyNodes {
		vs := VirtualServiceConfig{
			Name:        fmt.Sprintf("decoy-%s-%d", node, time.Now().Unix()),
			TargetHost:  topo.Nodes[node].Data.Service,
			MirrorRatio: 0.2 + rand.Float32()*0.3,
			FakePaths:   generateFakePaths(),
		}
		config.VirtualServices = append(config.VirtualServices, vs)
	}

	// 根据关键路径生成流量规则
	for _, path := range keyPaths {
		rule := TrafficRule{
			HeaderKey:   "X-Obfuscated-Trail",
			HeaderValue: generateTrailHash(path.Nodes),
			Probability: 0.1 + rand.Float32()*0.2,
		}
		config.TrafficRules = append(config.TrafficRules, rule)

		decoy := DecoyPath{
			Path:          mutatePath(path.Nodes),
			HoneypotCount: len(path.Nodes)/2 + 1,
		}
		config.DecoyPaths = append(config.DecoyPaths, decoy)
	}

	return config
}

// 部署虚拟服务的Kubernetes Operator
type K8sDeployer struct {
	Kubeconfig string
}

func (k *K8sDeployer) ApplyVirtualService(config VirtualServiceConfig) error {
	// 使用模板生成VirtualService配置
	const vsTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
name: {{.Name}}
spec:
hosts:
- "{{.TargetHost}}"
http:
- mirror:
		host: {{.Name}}-mirror
	mirrorPercentage: 
		value: {{.MirrorRatio}}
	route:
	- destination:
			host: {{.TargetHost}}
- match:
	- uri:
			prefix: /healthz
	fault:
		abort:
			percentage: 100
			httpStatus: 503
`

	tpl := template.Must(template.New("vs").Parse(vsTemplate))
	var buf bytes.Buffer
	tpl.Execute(&buf, config)

	// 调用kubectl应用配置
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = &buf
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("部署失败: %s, 输出: %s", err, string(output))
	}
	return nil
}

// 流量注入器
type TrafficInjector struct {
	HttpClient *http.Client
	Markers    []TrafficRule
}

func (t *TrafficInjector) StartInjection(targetURL string) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			for _, marker := range t.Markers {
				if rand.Float32() < marker.Probability {
					req, _ := http.NewRequest("GET", targetURL, nil)
					req.Header.Set(marker.HeaderKey, marker.HeaderValue)
					// 添加隐蔽标记
					req.Header.Set("X-Timestamp", fmt.Sprintf("%d", time.Now().UnixNano()^0xDEADBEEF))
					t.HttpClient.Do(req)
				}
			}
		}
	}
}

// Sidecar过滤器配置生成
func GenerateEnvoyFilter(config ObfuscationConfig) string {
	filterTemplate := `
name: obfuscation-filter
configPatches:
- applyTo: HTTP_FILTER
match:
	listener:
		filterChain:
			filter:
				name: envoy.filters.network.http_connection_manager
patch:
	operation: INSERT_BEFORE
	value:
		name: obfuscation-filter
		typed_config:
			"@type": type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua
			inlineCode: |
				function envoy_on_request(request_handle)
					local headers = request_handle:headers()
					{{ range .TrafficRules }}
					if headers:get("{{.HeaderKey}}") == "{{.HeaderValue}}" then
						request_handle:logInfo("Dropping obfuscated traffic")
						request_handle:respond(
							{[":status"] = "204"},
							"")
					end
					{{ end }}
				end
`
	tpl := template.Must(template.New("filter").Parse(filterTemplate))
	var buf bytes.Buffer
	tpl.Execute(&buf, config)
	return buf.String()
}

// 辅助函数
func generateFakePaths() []string {
	techs := []string{"api", "admin", "graphql", "health"}
	return []string{
		fmt.Sprintf("/%s/v1/%x", techs[rand.Intn(len(techs))], rand.Int31()),
		fmt.Sprintf("/%s/%d", techs[rand.Intn(len(techs))], time.Now().Unix()),
	}
}

func generateTrailHash(nodes []string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(nodes, ":"))))
}

func mutatePath(original []string) []string {
	mutated := make([]string, len(original))
	copy(mutated, original)
	if len(mutated) > 2 {
		mutated[1], mutated[len(mutated)-2] = mutated[len(mutated)-2], mutated[1]
	}
	return mutated
}

func calculateDegrees(t criticalpath.Topo) map[string]int {
	degrees := make(map[string]int)
	for _, edge := range t.Edges {
		degrees[edge.Data.Source]++
		degrees[edge.Data.Target]++
	}
	return degrees
}
func findKeyNodes(degrees map[string]int) []string {
	maxDegree := 0
	for _, d := range degrees {
		if d > maxDegree {
			maxDegree = d
		}
	}

	var keyNodes []string
	for node, d := range degrees {
		if d == maxDegree {
			keyNodes = append(keyNodes, node)
		}
	}
	return keyNodes
}
func findKeyPaths(allPaths []PathInfo) []PathInfo {
	if len(allPaths) == 0 {
		return nil
	}

	// 找最大度数总和
	maxSum := 0
	for _, path := range allPaths {
		if path.SumDegree > maxSum {
			maxSum = path.SumDegree
		}
	}

	// 筛选候选路径
	var candidates []PathInfo
	for _, path := range allPaths {
		if path.SumDegree == maxSum {
			candidates = append(candidates, path)
		}
	}

	// 找最大关键节点数
	maxKeyCount := 0
	for _, path := range candidates {
		if path.KeyNodeCount > maxKeyCount {
			maxKeyCount = path.KeyNodeCount
		}
	}

	// 最终筛选
	var keyPaths []PathInfo
	for _, path := range candidates {
		if path.KeyNodeCount == maxKeyCount {
			keyPaths = append(keyPaths, path)
		}
	}
	return keyPaths
}
func findAllPaths(topo criticalpath.Topo, currentPath []string, currentSum int, currentKeyCount int, visited map[string]bool, paths *[]PathInfo, degrees map[string]int, keyNodes map[string]bool) {
	currentNode := currentPath[len(currentPath)-1]

	// 记录有效路径（至少包含两个节点）
	if len(currentPath) >= 2 {
		*paths = append(*paths, PathInfo{
			Nodes:        append([]string{}, currentPath...),
			SumDegree:    currentSum,
			KeyNodeCount: currentKeyCount,
		})
	}

	// 遍历出边
	for _, edge := range topo.Edges {
		if edge.Data.Source == currentNode && !visited[edge.Data.Target] {
			newVisited := make(map[string]bool)
			for k, v := range visited {
				newVisited[k] = v
			}
			newVisited[edge.Data.Target] = true

			newPath := append(currentPath, edge.Data.Target)
			newSum := currentSum + degrees[edge.Data.Target]
			newKeyCount := currentKeyCount
			if keyNodes[edge.Data.Target] {
				newKeyCount++
			}

			findAllPaths(topo, newPath, newSum, newKeyCount, newVisited, paths, degrees, keyNodes)
		}
	}
}

func generateAllPaths(topo criticalpath.Topo, degrees map[string]int, keyNodes map[string]bool) []PathInfo {
	var paths []PathInfo
	for nodeID := range topo.Nodes {
		visited := make(map[string]bool)
		visited[nodeID] = true
		findAllPaths(topo,
			[]string{nodeID},
			degrees[nodeID],
			boolToInt(keyNodes[nodeID]),
			visited,
			&paths,
			degrees,
			keyNodes)
	}
	return paths
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func main() {
	// 1. 获取原始拓扑
	root := topo.GetTopo()
	shaped := criticalpath.shaping(root)

	// 2. 计算关键节点和路径
	degrees := calculateDegrees(shaped)
	keyNodes := findKeyNodes(degrees)
	keyPaths := findKeyPaths(generateAllPaths(shaped, degrees, keyNodes))

	// 3. 生成混淆策略
	config := GenerateObfuscation(shaped, keyNodes, keyPaths)

	// 4. 部署虚拟服务
	deployer := &K8sDeployer{Kubeconfig: "~/.kube/config"}
	for _, vs := range config.VirtualServices {
		if err := deployer.ApplyVirtualService(vs); err != nil {
			log.Printf("虚拟服务部署失败: %v", err)
		}
	}

	// 5. 应用Envoy过滤器
	filterConfig := GenerateEnvoyFilter(config)
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(filterConfig)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("过滤器部署失败: %s", string(output))
	}

	// 6. 启动流量注入
	injector := &TrafficInjector{
		HttpClient: &http.Client{Timeout: 5 * time.Second},
		Markers:    config.TrafficRules,
	}
	go injector.StartInjection("http://frontend.sockshop.svc.cluster.local")

	// 保持运行
	select {}
}
