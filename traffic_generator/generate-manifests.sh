#!/bin/bash

declare -A CONFIG_MAP=(
  ["instance1"]='["svc1.ns.svc.cluster.local","svc2.ns.svc.cluster.local"]'
  ["instance2"]='["svc3.ns.svc.cluster.local","svc4.ns.svc.cluster.local"]'
)

for instance in "${!CONFIG_MAP[@]}"; do
  # 生成Deployment
  sed -e "s/{{ .INSTANCE_ID }}/$instance/g" \
      -e "s/{{ .TARGET_SERVICES }}/${CONFIG_MAP[$instance]}/g" \
      deployment-template.yaml > "deploy-$instance.yaml"
  
  # 生成Service
  sed "s/{{ .INSTANCE_ID }}/$instance/g" \
      service-template.yaml > "svc-$instance.yaml"
done