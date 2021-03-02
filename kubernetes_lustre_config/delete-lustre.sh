#!/bin/sh

# Delete any client specific components
kubectl delete -f yaml/centos7-lustre-client.yaml
kubectl delete secret vmi-lustre-client-secret

# Delete any oss specific components

# The following are not always needed, should be uncommented only if the script did
# not work without them. Their purpose is to disable the finalizers of the resources
# which work like guard disallowing their deletion
# kubectl patch pvc vol-oss1 -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
# kubectl patch pvc vol-oss2 -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
# kubectl patch pv pv-oss1 -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
# kubectl patch pv pv-oss2 -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge

kubectl delete -f yaml/centos7-lustre-oss.yaml
kubectl delete -f yaml/pvc-oss2.yaml
kubectl delete -f yaml/pvc-oss1.yaml
kubectl delete -f yaml/pv-oss2.yaml
kubectl delete -f yaml/pv-oss1.yaml
kubectl delete secret vmi-lustre-oss-secret

# Delete any mds specific components
# kubectl patch pvc vol-mds -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
# kubectl patch pvc pv-mds -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
kubectl delete -f yaml/centos7-lustre-mds.yaml
kubectl delete -f yaml/pvc-mds.yaml
kubectl delete -f yaml/pv-mds.yaml
kubectl delete secret vmi-lustre-mds-secret

# Delete any mgs specific components
# kubectl patch pvc vol-mgs -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
# kubectl patch pvc pv-mgs -p ‘{“metadata”:{“finalizers”: []}}’ --type=merge
kubectl delete -f yaml/centos7-lustre-mgs.yaml
kubectl delete -f yaml/pvc-mgs.yaml
kubectl delete -f yaml/pv-mgs.yaml
kubectl delete secret vmi-lustre-mgs-secret