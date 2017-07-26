# crd-app-controller

Demonstrative implementation of Kubernetes Custom Resource Definitions (CRD).

This project provides a Custom Resource Definition for Apps, a few example App resources to deploy, and a controller to do something when resource events happen (create, delete, etc). The controller will create helm deployments for Drupal or WordPress applications as applicable.

## Usage

### Create the Custom Resource Definition

```
kubectl create -f app-resourcedef.yml
```

### Run the controller

The controller can be run from outside the cluster, which is handy for development.

```
go run *.go -kubeconfig=$HOME/.kube/config
```

### Create app resources

In another terminal, create app resources. Watch the output from the controller terminal window as you create or delete resources. You should see helm charts install for valid app types (wp and drupal), and you should see a failed deploy status for app-entry-invalid.yml.

```
kubectl create -f app-entry-[pickOne].yml
```

See your apps alongside other kube resources:

```
k get apps
```


```
k describe app [my-app-name]
```
