kubectl create namespace kubevirt
kubectl create configmap -n kubevirt kubevirt-config --from-literal debug.useEmulation=true
export RELEASE=v0.35.0
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/download/${RELEASE}/kubevirt-cr.yaml
