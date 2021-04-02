kubectl delete -f deploy/service_account.yaml
kubectl delete -f deploy/role.yaml
kubectl delete -f deploy/role_binding.yaml
kubectl delete -f deploy/crds/app_v1alpha1_podset_crd.yaml
kubectl delete -f deploy/crds/app_v1alpha1_podset_cr.yaml