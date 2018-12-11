# operator-framework usage for kubedge

## Initialization of kubedge-operator

```bash
git remote add origin git@github.com:kubedge/kubedge-operator
git push -u origin master
```

## Coding the kubedge-operator

```bash
operator-sdk add api --api-version=arpscan.kubedge.cloud/v1alpha1 --kind=Kubedge
vi pkg/apis/arpscan/v1alpha1/kubedge_types.go 
operator-sdk generate k8s
operator-sdk add controller --api-version=arpscan.kubedge.cloud/v1alpha1 --kind=Kubedge
cp /home/kubedge/proj/jbrette/operator-sdk/example/memcached-operator/memcached_controller.go.tmpl pkg/controller/kubedge/kubedge_controller.go
vi pkg/controller/kubedge/kubedge_controller.go
```


## Building the kubedge-operator

``bash
operator-sdk build kubedge/kubedge-operator-dev:v0.0.1
./manualbuild.sh 
```

## Deployment of kubedge-operator

```bash
kubectl create -f deploy/crds/arpscan_v1alpha1_kubedge_crd.yaml 
kubectl create -f deploy/service_account.yaml 
kubectl create -f deploy/role
kubectl create -f deploy/role.yaml 
kubectl create -f deploy/role_binding.yaml 
kubectl create -f deploy/operator.yaml 
```

## Deployment of arpscan

```bash
kubectl apply -f deploy/crds/arpscan_v1alpha1_kubedge_cr.yaml
```

```bash
kubectl get all

NAME                                    READY   STATUS    RESTARTS   AGE
pod/example-kubedge-5f9c4bfc95-hkbwk    1/1     Running   0          4m57s
pod/example-kubedge-5f9c4bfc95-qvctp    1/1     Running   0          4m57s
pod/example-kubedge-5f9c4bfc95-wjn9f    1/1     Running   0          4m57s
pod/kubedge-operator-5c89df997c-zclzh   1/1     Running   0          18m

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   30h

NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/example-kubedge    3/3     3            3           4m57s
deployment.apps/kubedge-operator   1/1     1            1           18m

NAME                                          DESIRED   CURRENT   READY   AGE
replicaset.apps/example-kubedge-5f9c4bfc95    3         3         3       4m57s
replicaset.apps/kubedge-operator-5c89df997c   1         1         1       18m
```

```bash
kubectl get kubedge/example-kubedge -o yaml

apiVersion: arpscan.kubedge.cloud/v1alpha1
kind: Kubedge
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"arpscan.kubedge.cloud/v1alpha1","kind":"Kubedge","metadata":{"annotations":{},"name":"example-kubedge","namespace":"default"},"spec":{"size":3}}
  creationTimestamp: "2018-12-11T20:20:50Z"
  generation: 2
  name: example-kubedge
  namespace: default
  resourceVersion: "41966"
  selfLink: /apis/arpscan.kubedge.cloud/v1alpha1/namespaces/default/kubedges/example-kubedge
  uid: 40beeaa7-fd82-11e8-bdcd-0800272318b1
spec:
  size: 3
  string: ""
status:
  nodes:
  - example-kubedge-5f9c4bfc95-hkbwk
  - example-kubedge-5f9c4bfc95-qvctp
  - example-kubedge-5f9c4bfc95-wjn9f
```
