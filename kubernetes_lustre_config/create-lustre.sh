#!/bin/sh

# This script assumes that the developer has already created an 
# ssh keypair and added it to the startup shell scripts. It also
# assumes that the user has created the directories in every cluster
# node that lustre will use to mount on. Finally, it assumes that the 
# yaml files used are adjusted. This mainly means that their requested
# storage size can be supported by the underlying host and that the
# nodeAffinity sections have the correct names of the cluster's nodes
# that will be used to deploy the pods

# Please contact aflpd@bu.edu if any of those is not clear before using the script

# This will enable deploying pods in the master node
kubectl taint nodes $(hostname) node-role.kubernetes.io/master:NoSchedule-

# Create any mgs specific components
kubectl create secret generic vmi-lustre-mgs-secret --from-file=userdata=lustre-mgs-startup.sh
kubectl create -f yaml/pv-mgs.yaml
kubectl create -f yaml/pvc-mgs.yaml
kubectl create -f yaml/centos7-lustre-mgs.yaml

# Create any mds specific components
kubectl create secret generic vmi-lustre-mds-secret --from-file=userdata=lustre-mds-startup.sh
kubectl create -f yaml/pv-mds.yaml
kubectl create -f yaml/pvc-mds.yaml
kubectl create -f yaml/centos7-lustre-mds.yaml

# Create any oss specific components
kubectl create secret generic vmi-lustre-oss-secret --from-file=userdata=lustre-oss-startup.sh
kubectl create -f yaml/pv-oss1.yaml
kubectl create -f yaml/pvc-oss1.yaml
kubectl create -f yaml/pv-oss2.yaml
kubectl create -f yaml/pvc-oss2.yaml
kubectl create -f yaml/centos7-lustre-oss.yaml

# Create any client specific components
kubectl create secret generic vmi-lustre-client-secret --from-file=userdata=lustre-client-startup.sh
kubectl create -f yaml/centos7-lustre-client.yaml