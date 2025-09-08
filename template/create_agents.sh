#!/bin/bash

# 检查是否传入了参数
if [ -z "$1" ]; then
  echo "请指定要创建的 test-agent 数量，例如： ./create_agents.sh 5"
  exit 1
fi

# 获取传入的参数数量
NUM_AGENTS=$1

# 创建指定数量的 test-agent 和 EnvoyFilter
for i in $(seq 1 $NUM_AGENTS); do
    # 创建 test-agent 部署
    sed "s/<instanceID>/$i/" test_agent.yaml | kubectl apply -f -

    # 创建 EnvoyFilter  
    sed "s/<instanceID>/$i/" test_service.yaml | kubectl apply -f -

    # 创建 EnvoyFilter
    sed "s/<instanceID>/$i/" test_envoy_filter.yaml | kubectl apply -f -

    echo "test-agent-$i 和 EnvoyFilter 已创建"
done