package obfuscation

import (
	"fmt"
	"math"

	criticalpath "github.com/sideChannel_topo_confusion/ce/criticalpath"
)

// 构建真实混淆链路
func BuildRealLinks(originalTopo criticalpath.Topo, criticalNodes []string, criticalPaths []criticalpath.PathInfo) []Link {
	var links []Link

	// 计算部署层数
	layerCount := int(math.Max(
		float64(len(criticalNodes)),
		float64(len(criticalPaths[0].Nodes)),
	))

	// 生成分层服务名称
	obfsServices := make([]string, layerCount)
	for i := 0; i < layerCount; i++ {
		obfsServices[i] = fmt.Sprintf("obfs-layer-%d", i+1)
	}

	// 构建链路规则
	for i, svc := range obfsServices {
		// 第一层连接入口
		if i == 0 {
			links = append(links, Link{
				Source: criticalPaths[0].Nodes[0], // 原入口服务
				Target: svc,
				Weight: 30, // 初始分流权重
			})
		}

		// 中间层连接规则
		if i > 0 && i < len(obfsServices)-1 {
			// 连接下层N-2个节点
			maxConnect := len(criticalNodes) - 2
			for j := 1; j <= maxConnect; j++ {
				if i+j < len(obfsServices) {
					links = append(links, Link{
						Source: svc,
						Target: obfsServices[i+j],
						Weight: 100 / (maxConnect + 1), // 动态权重分配
					})
				}
			}

			// 连接原关键路径后续节点
			if originalPos := i + 1; originalPos < len(criticalPaths[0].Nodes) {
				links = append(links, Link{
					Source: svc,
					Target: criticalPaths[0].Nodes[originalPos],
					Weight: 15, // 固定比例
				})
			}
		}

		// 最后一层不主动外连
	}

	return links
}

type Link struct {
	Source string
	Target string
	Weight int // 流量分配权重（百分比）
}
