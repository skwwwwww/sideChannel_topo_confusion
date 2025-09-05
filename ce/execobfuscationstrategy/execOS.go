package execobfuscationstrategy

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/sideChannel_topo_confusion/ce/criticalpath"
	generaltg "github.com/sideChannel_topo_confusion/ce/generateTG"
	generateobfucationstrategy "github.com/sideChannel_topo_confusion/ce/generateobfucationstrategy"
)

const (
	maxRetries            = 10               // 最大重试次数
	retryDelay            = 2 * time.Second  // 基础重试间隔
	recheckInterval       = 30 * time.Second // 重新评估策略的周期
	metricChangeThreshold = 0.1              // 流量特性变化触发重新混淆的阈值 (10%)
)

// 定义全局变量，用于存储上一次策略应用时的状态，以便进行比较和复用
var (
	currentPath                []string                              // 上一次的关键路径节点名称列表
	currentCriticalPathMetrics []criticalpath.CriticalPathNodeMetric // 上一次的关键路径节点指标
	currentNodesMap            map[string]criticalpath.TrafficNode   // 上一次的节点映射

	// 维护当前实际已部署的OA实例ID及其命名空间，用于管理OA的创建和删除
	// 格式: map[instanceID]namespace
	activeDeployedOAs = make(map[string]string)
	// 维护当前实际已部署的EnvoyFilter名称及其命名空间，用于管理EF的创建和删除
	// 格式: map[EnvoyFilterName]namespace
	activeDeployedEFs = make(map[string]string)
	rootEFConfig      generaltg.EnvoyFilterConfig // 记录根 EnvoyFilter 的配置 (App 和 Namespace
	rootEFName        string                      // 记录根 EnvoyFilter 的名称 (InstanceName
)

func Execobfuscationstrategy() {
	log.Println("这里初始化kubectl客户端,为之后创建OA和EnvoyFilter做准备")
	generaltg.InitClient()
	log.Println("初始化kubectl客户端成功")

	// 首次运行，强制部署混淆策略
	log.Println("首次部署混淆策略...")
	reapplyStrategy(true) // force = true

	// 进入无限循环，周期性地重新评估和应用策略
	for {
		log.Printf("等待 %s 后重新评估策略...\n", recheckInterval)
		time.Sleep(recheckInterval) // 等待指定时间
		log.Println("正在重新评估混淆策略...")
		reapplyStrategy(false) // force = false，进行变化检查
	}
}

func reapplyStrategy(force bool) {
	// 1. 获取最新的拓扑和关键路径信息
	_, newPath, newNodesMap, newCriticalPathMetrics := criticalpath.GetCriticalPaths()

	// 2. 根据最新的关键路径信息生成新的混淆策略配置
	newOAConfig, newEnvoyFilterConfig, newDownstreamNodeConfig := generateobfucationstrategy.Generateconfucationstrategy(newPath, newNodesMap, newCriticalPathMetrics)

	// 决定是否需要重新应用策略
	needsReapply := force // 如果是强制模式（例如首次运行），则直接重新应用

	if !force { // 非强制模式下，进行变化检测
		// 检查关键路径是否发生变化（节点或长度）
		pathChanged := !arePathsSame(currentPath, newPath)

		// 如果路径没有变化，则检查流量特性是否有显著变化
		metricsChanged := false
		if !pathChanged && currentCriticalPathMetrics != nil { // 只有路径相同且存在旧指标时才检查
			metricsChanged = areMetricsSignificantlyDifferent(currentCriticalPathMetrics, newCriticalPathMetrics, metricChangeThreshold)
		}

		needsReapply = pathChanged || metricsChanged // 只要路径或指标有变化，就需要重新应用

		if !needsReapply {
			log.Printf("未检测到关键路径或流量特性有显著变化，跳过重新应用策略。\n")
			return // 没有变化，直接返回
		}

		// 打印触发重新应用的具体原因
		if pathChanged {
			log.Printf("关键路径发生变化（节点或长度），正在重新应用混淆策略。\n")
		}
		if metricsChanged {
			log.Printf("流量特性发生显著变化，正在重新应用混淆策略。\n")
		}
	}

	// 执行到这里说明需要重新应用策略
	log.Println("正在应用新的混淆策略...")

	// --- OA 管理 (复用、创建、删除) ---
	desiredActiveOAs := make(map[string]struct{}) // 用于记录本次策略应用后期望活跃的OA实例

	// 遍历新的OA配置，处理OA的创建、复用和下游节点设置
	for i, newOACfg := range newOAConfig {
		// OA的InstanceID与其在关键路径中的索引（位置）绑定
		instanceID := fmt.Sprintf("%d", i)
		currentOANamespace, exists := activeDeployedOAs[instanceID] // 检查此ID的OA是否已部署

		if !exists || currentOANamespace != newOACfg.Namespace {
			// 如果OA不存在，或者其命名空间发生变化（不常见，但以防万一），则创建
			log.Printf("创建或确保OA %s 在命名空间 %s 中存在...\n", instanceID, newOACfg.Namespace)
			generaltg.CreateOA(newOACfg.Namespace, instanceID)
			activeDeployedOAs[instanceID] = newOACfg.Namespace // 标记为已部署
		} else {
			log.Printf("复用现有OA %s 在命名空间 %s 中。\n", instanceID, newOACfg.Namespace)
		}
		desiredActiveOAs[instanceID] = struct{}{} // 标记为本次期望活跃的OA

		// 设置/更新此OA的下游节点
		if i < len(newDownstreamNodeConfig) && len(newDownstreamNodeConfig[i]) > 0 {
			oaSourceDNS := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+instanceID, newOACfg.Namespace)
			log.Printf("为OA %s (%s) 设置下游节点...\n", instanceID, oaSourceDNS)
			generaltg.SetDownstreamNode(newDownstreamNodeConfig[i], oaSourceDNS)
		} else {
			log.Printf("未为OA %s 配置下游节点。跳过设置。\n", instanceID)
		}

		// 在设置下游节点后，对新创建/更新的OA进行健康检查
		serviceURL := "http://" + fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+instanceID, newOACfg.Namespace) + "/healthz"
		numRetries := 0
		for {
			err := checkService(serviceURL)
			if err == nil {
				log.Printf("服务 %s 健康检查通过。\n", serviceURL)
				break
			}
			if numRetries >= maxRetries {
				log.Printf("错误: 服务 %s 经过 %d 次重试后健康检查失败: %v\n", serviceURL, maxRetries, err)
				break // 超过最大重试次数，退出重试循环
			}
			delay := retryDelay * time.Duration(numRetries+1) // 线性退避
			log.Printf("服务 %s 不健康。将在 %s 后重试 (尝试 %d/%d)。\n", serviceURL, delay, numRetries+1, maxRetries)
			time.Sleep(delay)
			numRetries++
		}
	}
	// 清理不再是关键路径一部分的旧OA
	for instanceID, ns := range activeDeployedOAs {
		if _, exists := desiredActiveOAs[instanceID]; !exists {
			log.Printf("删除不再使用的OA %s 在命名空间 %s 中...\n", instanceID, ns)
			generaltg.DeleteOA(ns, instanceID) // 假设 generaltg.DeleteOA 函数存在
			delete(activeDeployedOAs, instanceID)
		}
	}

	// --- EnvoyFilter 管理 (创建、更新、删除) ---
	desiredActiveEFs := make(map[string]struct{}) // 用于记录本次策略应用后期望活跃的EnvoyFilter实例

	// 遍历新的 EnvoyFilter 配置（除根 EnvoyFilter 外）
	// 注意：这里假设 CreateEnvoyFilter 会处理更新逻辑（如果EF已存在则更新）
	for i := 0; i < len(newEnvoyFilterConfig)-1; i++ {
		efName := newEnvoyFilterConfig[i].InstanceName
		efNamespace := newEnvoyFilterConfig[i].Namespace

		currentEFNamespace, exists := activeDeployedEFs[efName]
		if !exists || currentEFNamespace != efNamespace {
			log.Printf("创建或确保 EnvoyFilter %s 在命名空间 %s 中存在...\n", efName, efNamespace)
			generaltg.CreateEnvoyFilter(efNamespace, efName)
			activeDeployedEFs[efName] = efNamespace
		} else {
			log.Printf("复用现有 EnvoyFilter %s 在命名空间 %s 中。\n", efName, efNamespace)
		}
		desiredActiveEFs[efName] = struct{}{} // 标记为本次期望活跃的EF
	}

	// 处理根 EnvoyFilter
	rootEnvoyFilterName := ""
	rootEnvoyFilterNamespace := ""
	if len(newEnvoyFilterConfig) > 0 && len(newOAConfig) > 0 {
		rootConfig := generaltg.EnvoyFilterConfig{
			Namespace: newEnvoyFilterConfig[len(newEnvoyFilterConfig)-1].Namespace,
			App:       newEnvoyFilterConfig[len(newEnvoyFilterConfig)-1].InstanceName,
		}
		rootEnvoyFilterName = "filter-confusion-header" + rootConfig.App // 假设根EnvoyFilter的命名规则和普通EnvoyFilter类似
		rootEnvoyFilterNamespace = rootConfig.Namespace

		// 根 EnvoyFilter 指向第一个OA
		rootTargetService := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+newOAConfig[0].InstanceID, newOAConfig[0].Namespace)

		currentRootEFNamespace, exists := activeDeployedEFs[rootEnvoyFilterName]
		if !exists || currentRootEFNamespace != rootEnvoyFilterNamespace {
			log.Printf("创建或确保根 EnvoyFilter (App: %s, Namespace: %s) 指向 %s 存在...\n", rootConfig.App, rootConfig.Namespace, rootTargetService)
			generaltg.CreateRootEnvoyFilter(rootConfig, rootTargetService)
			rootEFConfig = rootConfig // 更新全局记录
			rootEFName = rootEnvoyFilterName
			activeDeployedEFs[rootEnvoyFilterName] = rootEnvoyFilterNamespace
		} else {
			log.Printf("复用现有根 EnvoyFilter (App: %s, Namespace: %s) 指向 %s。\n", rootConfig.App, rootConfig.Namespace, rootTargetService)
		}
		desiredActiveEFs[rootEnvoyFilterName] = struct{}{}
	} else {
		log.Printf("尝试删除根 EnvoyFilter: %s/%s\n", rootEFConfig, rootEFName)
		generaltg.DeleteRootEnvoyFilter(rootEFConfig, rootEFName) // 假设 DeleteEnvoyFilter 可以通用删除所有 EF
	}

	// 清理不再是新策略一部分的旧 EnvoyFilter
	for efName, ns := range activeDeployedEFs {
		if _, exists := desiredActiveEFs[efName]; !exists {
			log.Printf("尝试删除普通 EnvoyFilter: %s/%s\n", ns, efName)
			generaltg.DeleteEnvoyFilter(ns, efName) // 调用通用的删除函数
			delete(activeDeployedEFs, efName)       // 从活跃列表中移除
		}
	}

	// --- 更新全局状态，为下一次循环做准备 ---
	currentPath = newPath
	currentNodesMap = newNodesMap
	currentCriticalPathMetrics = newCriticalPathMetrics

	log.Println("混淆策略重新应用完成。")
}

// arePathsSame 检查两条关键路径是否完全相同（长度和内容）
func arePathsSame(path1, path2 []string) bool {
	if len(path1) != len(path2) {
		return false
	}
	for i := range path1 {
		if path1[i] != path2[i] {
			return false
		}
	}
	return true
}

// areMetricsSignificantlyDifferent 检查流量指标是否显著不同（超过阈值）
func areMetricsSignificantlyDifferent(metrics1, metrics2 []criticalpath.CriticalPathNodeMetric, threshold float64) bool {
	if len(metrics1) != len(metrics2) {
		return true // 长度不同即视为显著差异
	}

	for i := range metrics1 {
		if metrics1[i].ServiceNum != metrics2[i].ServiceNum {
			return true
		}
		// 检查 RPS
		if metrics1[i].Rps > 0 && math.Abs((metrics1[i].Rps-metrics2[i].Rps)/metrics1[i].Rps) > threshold {
			return true
		} else if metrics1[i].Rps == 0 && metrics2[i].Rps > 0 {
			return true
		}
		// 检查 ErrorRate
		if metrics1[i].ErrorRate > 0 && math.Abs((metrics1[i].ErrorRate-metrics2[i].ErrorRate)/metrics1[i].ErrorRate) > threshold {
			return true
		} else if metrics1[i].ErrorRate == 0 && metrics2[i].ErrorRate > 0 {
			return true
		}
	}
	return false
}

// checkService 用来判断创建的OA是否已经可用了（如果不进行判断，直接设置downstreamNode会报错）
func checkService(serviceURL string) error {
	resp, err := http.Get(serviceURL)
	log.Println(resp)
	if resp == nil {
		log.Printf("警告: HTTP响应为空,服务可能不可用 %s", serviceURL)
		return fmt.Errorf("HTTP响应为空")
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("服务 %s 不可用，状态码: %d, 错误: %v", serviceURL, resp.StatusCode, err)
		return fmt.Errorf("服务不可用，状态码: %d, 错误: %v", resp.StatusCode, err)
	}
	return nil
}
