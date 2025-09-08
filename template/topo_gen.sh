#!/bin/bash

# 获取传入的 agent 数量 n
n=$1

# 计算每层的 agent 数量
if [ $((n / 3)) -lt 10 ]; then
    agents_per_layer=$n
else
    agents_per_layer=10
fi

# 初始化当前 agent 编号
current_agent=1

# 设置下游节点
while [ $current_agent -le $n ]; do
    # 计算当前 agent 的下游节点
    downstream_nodes=()

    # 下游节点是当前节点的下一个若干个节点，最多 10 个
    for ((i=1; i<=agents_per_layer; i++)); do
        downstream_node=$((current_agent + i))
        if [ $downstream_node -le $n ]; then
            downstream_nodes+=("$downstream_node")
        fi
    done

    # 创建 JSON 数据，设置下游节点
    downstream_json="["
    for node in "${downstream_nodes[@]}"; do
        downstream_json+="{\"DNS\": \"test-agent-$node.default.svc.cluster.local:80/api\", \"ServiceNum\": 1, \"Rps\": 0.465, \"ErrorRate\": 0},"
    done
    downstream_json="${downstream_json%,}]"

    # 使用 curl 设置下游节点
    curl -X POST http://test-agent-$current_agent.default.svc.cluster.local/set-nodes -d "$downstream_json"

    # 移动到下一个 agent
    current_agent=$((current_agent + 1))
done
