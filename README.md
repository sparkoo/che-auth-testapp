# Che authentication / authorization test app

This is test application for prototyping of new ways of authentication / authorization in Che.

### Image
```
quay.io/mvala/che-auth-testapp:latest
```
#### Build
```
$ ./build.sh
```

### How it works
App is simple web-server. 
It reads the Authorization bearer token from the request header, and uses the token to configure new kubernetes client. 
It then reads the request path and uses it as a namespace to query several k8s objects.

### How to use is
#### Simple minikube setup
 1. `minikube.sh` script will start new minikube instance with 5 users `user[1-5]` (defined in `minikube_users.csv`).
 2. `kc apply`
    1. `01_namespaces.yaml` - namespaces for the users `user[1-5]-ns`
    2. `02_rbac.yaml` - admin roles for the users only to their namespace 
    3. `03_deployment.yaml` - deployment of this test app to `che` namespaces
 3. `test.sh <URL> <NAMESPACE> <TOKEN>` - test script to test the setup

### Exmaple
#### Successful authentication and authorization
request to `user1-ns` namespace with `user1`'s token:
```
[~/dev/che-auth-testapp] λ ./test.sh che.192.168.39.78.nip.io user1-ns user1-token
Hi there, Try to get resources from [user1-ns] namespace.
Using authorization bearer token [user1-token]

ConfigMaps
========
 - kube-root-ca.crt
Secrets
========
Pods
========
```

#### Successful authentication but unsuccessful authorization 
request to `user1-ns` namespace with `user2`'s token:
```
[~/dev/che-auth-testapp] λ ./test.sh che.192.168.39.78.nip.io user1-ns user2-token
Hi there, Try to get resources from [user1-ns] namespace.
Using authorization bearer token [user2-token]

Something went wrong. I can't get the configMaps. [configmaps is forbidden: User "user2" cannot list resource "configmaps" in API group "" in the namespace "user1-ns"]
Something went wrong. I can't get the secrets. [pods is forbidden: User "user2" cannot list resource "pods" in API group "" in the namespace "user1-ns"]
Something went wrong. I can't get the pods. [pods is forbidden: User "user2" cannot list resource "pods" in API group "" in the namespace "user1-ns"]
```

#### Unsuccessful authentication
try to use unknown token:
```
[~/dev/che-auth-testapp] λ ./test.sh che.192.168.39.78.nip.io user1-ns dont-know  
Hi there, Try to get resources from [user1-ns] namespace.
Using authorization bearer token [dont-know]

Something went wrong. I can't get the configMaps. [Unauthorized]
Something went wrong. I can't get the secrets. [Unauthorized]
Something went wrong. I can't get the pods. [Unauthorized]
```
