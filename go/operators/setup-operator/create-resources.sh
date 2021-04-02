kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/app_v1alpha1_podset_crd.yaml
kubectl create -f deploy/crds/app_v1alpha1_podset_cr.yaml