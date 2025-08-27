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

	// oaConfig, envoyFilterConfig, downstreamNodeConfig := generateobfucationstrategy.Generateconfucationstrategy()
	// for i := 0; i < len(oaConfig); i++ {
	// 	generaltg.CreateOA(oaConfig[i].Namespace, oaConfig[i].InstanceID)
	// }
	// for i := 0; i < len(envoyFilterConfig)-1; i++ {
	// 	generaltg.CreateEnvoyFilter(envoyFilterConfig[i].Namespace, envoyFilterConfig[i].InstanceName)
	// }
	// if len(envoyFilterConfig) > 0 && len(oaConfig) > 0 {
	// 	config := generaltg.EnvoyFilterConfig{
	// 		Namespace: (envoyFilterConfig[len(envoyFilterConfig)-1].Namespace),
	// 		App:       envoyFilterConfig[len(envoyFilterConfig)-1].InstanceName,
	// 	}

	// 	generaltg.CreateRootEnvoyFilter(config, fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+oaConfig[0].InstanceID, oaConfig[0].Namespace))
	// }

	// // 这里需要修改OA和， 最后一层不用设置，所以-1
	// for i := 0; i < len(downstreamNodeConfig)-1; i++ {

	// 	DNS := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+oaConfig[i].InstanceID, oaConfig[i].Namespace)
	// 	serviceURL := "http://" + DNS + "/healthz"
	// 	fmt.Println(serviceURL)
	// 	num := 0
	// 	for {
	// 		err := checkService(serviceURL)
	// 		if err == nil {
	// 			break
	// 		}

	// 		if num >= maxRetries {
	// 			fmt.Errorf("超过最大重试次数%d次", maxRetries)
	// 		}

	// 		delay := retryDelay * time.Duration(i+1) // 线性退避
	// 		fmt.Printf("第%d次重试，等待%s后重试\n", num, delay)
	// 		time.Sleep(delay)
	// 		num++
	// 	}
	// 	generaltg.SetDownstreamNode(downstreamNodeConfig[i], DNS)

	// }

}

func reapplyStrategy(force bool) {
	// 1. 获取最新的拓扑和关键路径信息
	_, newPath, newNodesMap, newCriticalPathMetrics := criticalpath.GetCriticalPaths()

	// 2. 根据最新的关键路径信息生成新的混淆策略配置
	// `Generateconfucationstrategy` 函数已被修改，现在接收这些参数
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

	// --- OA 管理 (复用、创建、删除) ---
	// 用于记录本次策略应用后期望活跃的OA实例
	desiredActiveOAs := make(map[string]struct{})

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

		// todo 感觉这里不太对
		// 设置/更新此OA的下游节点
		// 检查新的下游节点配置是否存在此OA的配置
		if i < len(newDownstreamNodeConfig) && len(newDownstreamNodeConfig[i]) > 0 {
			// 根据OA的InstanceID和Namespace构建其DNS名称，作为SetDownstreamNode的源头参数
			oaSourceDNS := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+instanceID, newOACfg.Namespace)
			log.Printf("为OA %s (%s) 设置下游节点...\n", instanceID, oaSourceDNS)
			// generaltg.SetDownstreamNode 期望接收一个 DownstreamNodeConfig 切片
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
			generaltg.DeleteOA(ns, instanceID)    // 假设 generaltg.DeleteOA 函数存在
			delete(activeDeployedOAs, instanceID) // 从活跃列表中移除
		}
	}
	// --- EnvoyFilter 管理 ---
	// EnvoyFilter 的更新通常可以通过重新应用来完成，Istio 会进行协调。
	// 这里直接基于新的配置进行创建/更新操作。
	// 遍历新的 EnvoyFilter 配置（除根 EnvoyFilter 外）
	for i := 0; i < len(newEnvoyFilterConfig)-1; i++ {
		log.Printf("创建/更新 EnvoyFilter %s 在命名空间 %s 中...\n", newEnvoyFilterConfig[i].InstanceName, newEnvoyFilterConfig[i].Namespace)
		generaltg.CreateEnvoyFilter(newEnvoyFilterConfig[i].Namespace, newEnvoyFilterConfig[i].InstanceName)
	}

	// 处理根 EnvoyFilter
	if len(newEnvoyFilterConfig) > 0 && len(newOAConfig) > 0 {
		// 根 EnvoyFilter 通常是 newEnvoyFilterConfig 的最后一个元素
		rootConfig := generaltg.EnvoyFilterConfig{
			Namespace: newEnvoyFilterConfig[len(newEnvoyFilterConfig)-1].Namespace,
			App:       newEnvoyFilterConfig[len(newEnvoyFilterConfig)-1].InstanceName,
		}
		// 根 EnvoyFilter 指向第一个OA
		rootTargetService := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+newOAConfig[0].InstanceID, newOAConfig[0].Namespace)
		log.Printf("创建/更新根 EnvoyFilter (App: %s, Namespace: %s) 指向 %s...\n", rootConfig.App, rootConfig.Namespace, rootTargetService)
		generaltg.CreateRootEnvoyFilter(rootConfig, rootTargetService)
	} else {
		log.Println("跳过根 EnvoyFilter 创建：配置不足。")
	}

	// --- 更新全局状态，为下一次循环做准备 ---
	currentPath = newPath
	currentNodesMap = newNodesMap
	currentCriticalPathMetrics = newCriticalPathMetrics
	// currentOAConfigs 只是存储生成的配置，activeDeployedOAs 才是实际部署的OA状态
	// currentOAConfigs = newOAConfig // 这一行在实际决策中不直接使用，activeDeployedOAs 更重要

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
		// 如果长度不同，通常意味着路径结构已变，应由 `arePathsSame` 捕获。
		// 如果到达这里，可能是数据不一致，视为显著差异。
		return true
	}

	for i := range metrics1 {
		// 检查 ServiceNum (服务数量通常是整数，任何变化都视为变化)
		if metrics1[i].ServiceNum != metrics2[i].ServiceNum {
			return true
		}
		// 检查 RPS (请求率)
		// 避免除以零：如果旧RPS为0，而新RPS大于0，则视为变化。
		if metrics1[i].Rps > 0 && math.Abs((metrics1[i].Rps-metrics2[i].Rps)/metrics1[i].Rps) > threshold {
			return true
		} else if metrics1[i].Rps == 0 && metrics2[i].Rps > 0 {
			return true
		}
		// 检查 ErrorRate (错误率)
		// 避免除以零：如果旧错误率为0，而新错误率大于0，则视为变化。
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
		log.Fatalf("响应为空")
		return err
	}
	if err != nil || resp.StatusCode != 200 {
		log.Fatalf("服务不可用，状态码: %d", resp.StatusCode)
		return err
	}
	return nil
}
