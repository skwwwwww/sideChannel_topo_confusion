#!/bin/bash

# 检查是否传入了参数
if [ -z "$1" ]; then
  echo "请指定要删除的 test-agent 数量，例如： ./delete_agents.sh 5"
  exit 1
fi

# 获取传入的参数数量
NUM_AGENTS=$1

# 删除指定数量的 test-agent 和 EnvoyFilter
for i in $(seq 1 $NUM_AGENTS); do
    # 删除 test-agent 部署
    kubectl delete deploy test-agent-$i

    # 删除 test-agent 服务
    kubectl delete svc test-agent-$i

    # 删除 EnvoyFilter
    kubectl delete envoyfilter filter-confusion-header-$i

    echo "test-agent-$i 和 EnvoyFilter 已删除"
done
