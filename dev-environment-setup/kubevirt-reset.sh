#!/bin/sh
kubectl delete secrets --namespace kubevirt -l kubevirt.io
kubectl delete pods --namespace kubevirt -l kubevirt.io
