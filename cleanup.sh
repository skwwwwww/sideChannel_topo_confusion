#!/bin/bash

# --- Kubernetes Resource Cleanup Script ---

echo "Deleting Pod 'ce'..."
kubectl delete po ce --ignore-not-found=true

echo "Deleting Deployments: traffic-service-0, traffic-service-1, traffic-service-2..."
kubectl delete deploy traffic-service-0 traffic-service-1 traffic-service-2 --ignore-not-found=true

echo "Deleting Services: traffic-service-0, traffic-service-1, traffic-service-2..."
kubectl delete svc traffic-service-0 traffic-service-1 traffic-service-2 --ignore-not-found=true

echo "Deleting EnvoyFilters..."
kubectl delete envoyfilter \
    filter-confusion-headerratings \
    filter-confusion-headerreviews \
    filter-confusion-headertraffic-service-0 \
    filter-confusion-headertraffic-service-1 \
    filter-confusion-headertraffic-service-2 \
    filter-confusion-header-root \
    --ignore-not-found=true

echo "Cleanup commands executed."
