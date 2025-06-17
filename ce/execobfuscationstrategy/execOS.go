package execobfuscationstrategy

import (
	"fmt"
	"net/http"
	"time"

	generaltg "github.com/sideChannel_topo_confusion/ce/generateTG"
	generateobfucationstrategy "github.com/sideChannel_topo_confusion/ce/generateobfucationstrategy"
)

const (
	maxRetries = 10              // 最大重试次数
	retryDelay = 3 * time.Second // 基础重试间隔
)

func Execobfuscationstrategy() {
	generaltg.InitClient()
	oaConfig, envoyFilterConfig, downstreamNodeConfig := generateobfucationstrategy.Generateconfucationstrategy()
	for i := 0; i < len(oaConfig); i++ {
		generaltg.CreateOA(oaConfig[i].Namespace, oaConfig[i].InstanceID)
	}
	for i := 0; i < len(envoyFilterConfig)-1; i++ {
		generaltg.CreateEnvoyFilter(envoyFilterConfig[i].Namespace, envoyFilterConfig[i].InstanceName)
	}
	if len(envoyFilterConfig) > 0 && len(oaConfig) > 0 {
		config := generaltg.EnvoyFilterConfig{
			Namespace: (envoyFilterConfig[len(envoyFilterConfig)-1].Namespace),
			App:       envoyFilterConfig[len(envoyFilterConfig)-1].InstanceName,
		}

		generaltg.CreateRootEnvoyFilter(config, fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+oaConfig[0].InstanceID, oaConfig[0].Namespace))
	}

	// 这里需要修改OA和， 最后一层不用设置，所以-1
	for i := 0; i < len(downstreamNodeConfig)-1; i++ {

		DNS := fmt.Sprintf("%s.%s.svc.cluster.local", "traffic-service-"+oaConfig[i].InstanceID, oaConfig[i].Namespace)
		serviceURL := "http://" + DNS + "/healthz"
		fmt.Println(serviceURL)
		num := 0
		for {
			err := checkService(serviceURL)
			if err == nil {
				break
			}

			if num >= maxRetries {
				fmt.Errorf("超过最大重试次数%d次", maxRetries)
			}

			delay := retryDelay * time.Duration(i+1) // 线性退避
			fmt.Printf("第%d次重试，等待%s后重试\n", num, delay)
			time.Sleep(delay)
			num++
		}
		generaltg.SetDownstreamNode(downstreamNodeConfig[i], DNS)

	}

}

func checkService(serviceURL string) error {

	resp, err := http.Get(serviceURL)
	fmt.Println(resp)
	if resp == nil {
		return fmt.Errorf("响应为空")
	}
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("服务不可用，状态码: %d", resp.StatusCode)
	}
	return nil
}
