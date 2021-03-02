#!/bin/sh
kubeadm init --apiserver-advertise-address=10.0.1.233 --pod-network-cidr=10.244.0.0/16
