## Flux Secret Operator
This repository implements a custom [controller](https://github.com/mendoncangelo/flux-secret/blob/master/pkg/controller/stackoverdrivesecret/stackoverdrivesecret_controller.go) for watching StackOverDriveSecret resources as defined with a [CustomResourceDefinition (CRD)](https://github.com/mendoncangelo/flux-secret/blob/master/config/crds/fluxsecret_v1beta1_stackoverdrivesecret.yaml).

### Overview
The StackOverDriveSecret can encompass one or more secrets in a specific namespace. The CRD has been NameSpaced scoped. Meaning that the CR instances will need a NameSpace to exist in as opposed to Cluster scoped where the instance of the CRD can exist at the Cluster scope. The corresponding secrets will be created in the same Namespace. This namespace needs to be created prior to creating the Custom Resource Instance. A typical yaml definition of a StackOverDriveSecret would look like [this](https://github.com/mendoncangelo/flux-secret/blob/master/config/samples/fluxsecret_v1beta1_stackoverdrivesecret.yaml)

### Prerequisites
You need to have Golang installed and dep for managing dependencies. We need to use Kubernetes 1.7 or greater as Custom Resources are supported in version 1.7 and above. Dependencies are managed via [dep](https://github.com/golang/dep) since kubebuilder 1.0.8 uses dep for managing Go dependencies. The vendor folder has been committed. So even if the network is down you can still build successfully. You also need [kustomize](https://sigs.k8s.io/kustomize) installed. 

### Install Go

StackOverDriveSecret Operator is written in Go.  On OS X, use Homebrew:

```shell
brew install go
# Make sure to set the GOPATH
```
For other operating systems see [Installing Go](https://golang.org/doc/install).

### Mini K8s Cluster for Testing
You can choose to test on any k8s cluster. I have chosen to test on [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/).

### Get the Source Code:
```bash
# If you want to clone inside the GOPATH(GOPATH environment variable should be set)
go get -d https://github.com/mendoncangelo/flux-secret
```

### Build and Testing Locally:
```bash
# To build the code
make build
```
To make Kubernetes accept your custom StackOverDriveSecret resource instances, you need to post the StackOverDriveSecret CustomResourceDefinition to the API server. After you post the descriptor to Kubernetes, it will allow you to create any number of instances of the custom StackOverDriveSecret resource.

```bash
# To Install the Custom Resource Definition into the cluster. You need to set the KUBECONFIG variable for kubectl to work.
make install
```
Now that a StackOverDriveSecret object has been created, you can now store, retrieve, and delete Custom StackOverDriveSecret Objects through the Kubernetes API
server. These objects don’t do anything yet. That is where the custom controller comes into play. The controller needs to be deployed 

For Local Development and Testing you can run the main.go file in the `cmd/manager` dir. The command to run the file is 
```bash
KUBECONFIG=~/.kube/config go run cmd/manager/main.go
```
You need to set the KUBECONFIG environment variable to point to your kubeconfig file. That will run the controller locally instead of packaging into a container and deploying it as a Pod. However, for systems other than local development it is recommended to deploy it as a StatefulSet in the Kubernetes cluster. 


### Changing the Custom Resource Definition:
If you are actively developing and making changes to the CustomResourceDefinition, your latest changes will not be reflected inside of the Kubernetes API Server unless you dont generate and apply the new CRD's.

```bash
# Install CRDs into a cluster. This will also generates new manifests
make install.
```

### TODO: Running Unit Tests.
Not yet implemented. Need to write tests.
```bash
# make test
```

### Deploying the StackOverDriveSecret Custom Controller
```bash
# This will deploy the controller in the configured cluster in ~/.kube/config
# The controller will start but it won't do anything just yet. We need to apply a StackOverDriveSecret Sample which will create a StackOverDriveSecret Custom Resource. I am using kustomize to manage the yaml files. You can run 
kustomize build config/default | kubectl apply -f.
# This will create all the necessary resources for the custom controller. 
make deploy
```

### Create Custom Resource
```bash
# This will create Custom Resources in the cluster. Make sure to edit the sample templates if you've changed the API definition. 
make sample-deployment
# or you can directly run 
kubectl apply -f config/samples/fluxsecret_v1beta1_stackoverdrivesecret.yaml
```

<!-- The secret can be encrypted using an aws kms key.  -->
```bash
aws --region REGION_IN_WHICH_KEY_EXIST kms encrypt --key-id YOUR_KEY_ID --plaintext file://PRIVATE_KEY_LOCATION --output text --query CiphertextBlob
```
Take the resulting text and put in a file like [this](https://github.com/mendoncangelo/flux-secret/blob/master/config/samples/fluxsecret_v1beta1_stackoverdrivesecret.yaml#L10). 

### Example Run for Development
I used [minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/#installation) for testing the controller. The controller when invoked will read the `encryptedData` inside of the custom resource instance, base64 decode it and then decrypt it using AWS KMS. Before you start make sure you have an appropriate IAM policy created in  AWS IAM that permits the user to use the KMS service. KUBECONFIG should also be set before you run kubectl commands. 

When we build the docker image the image should be built with the context of the docker enginer running inside the
minikube vm. First, set the environment using `eval $(minikube docker-env)` from your localhost. In a production scenario you would build, tag and push the docker image to a docker registry. e.g. AWS ECR and then apply the yaml definition via flux.

Run the `make docker-build` command. This should build the docker image in the context of the docker engine. If you login to the minikube vm `minikube ssh` and execute `docker ps | grep controller` you will see the latest controller image built. 
After you have a docker image built make sure to add the AWS credentials [AWS_ACCESS_KEY_ID](https://github.com/mendoncangelo/flux-secret/blob/master/config/default/manager/manager.yaml#L59) and [AWS_SECRET_ACCESS_KEY](https://github.com/mendoncangelo/flux-secret/blob/master/config/default/manager/manager.yaml#L61) to the manager.yaml file to pass into the container. If you dont do this the pod will not be able to authenticate to aws and use the kms key. 

For the sake of this example solution I have chosen to pass in aws credentials into the Pod via environment variables. In a production scenario I would choose to use kube2iam to get a specific IAM role for the container so that it can use the kms keys that were used to encrypt the data to decrypt them. The controller takes the encrypted data, base64 decodes it as it was not base64 decoded in the first step and creates a Secret in the same namespace as the namespace the controller is running in. The name of the secret is the key that is defined in the yaml file. e.g. ssh-privatekey

Once that is done you can run `kustomize build config/default | kubectl apply -f -`. This will deploy all
the resources and rbac's into the k8s cluster. You can open the minikube dashboard(needs to be installed if
you dont have it) and check in the flex-secrets-system namespace. If you look at the logs of the manager in the
dashboard you should see something like this
```
{"level":"info","ts":1564671398.5742247,"logger":"entrypoint","msg":"setting up client for manager"}
{"level":"info","ts":1564671398.5745244,"logger":"entrypoint","msg":"setting up manager"}
{"level":"info","ts":1564671398.8612962,"logger":"entrypoint","msg":"Registering Components."}
{"level":"info","ts":1564671398.8613412,"logger":"entrypoint","msg":"setting up scheme"}
{"level":"info","ts":1564671398.8616164,"logger":"entrypoint","msg":"Setting up controller"}
{"level":"info","ts":1564671398.8617764,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"stackoverdrivesecret-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1564671398.861902,"logger":"kubebuilder.controller","msg":"Starting EventSource","controller":"stackoverdrivesecret-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1564671398.8622308,"logger":"entrypoint","msg":"setting up webhooks"}
{"level":"info","ts":1564671398.8622625,"logger":"entrypoint","msg":"Starting the Cmd."}
{"level":"info","ts":1564671398.9637914,"logger":"kubebuilder.controller","msg":"Starting Controller","controller":"stackoverdrivesecret-controller"}
{"level":"info","ts":1564671399.0645976,"logger":"kubebuilder.controller","msg":"Starting workers","controller":"stackoverdrivesecret-controller","worker count":1}
```
The custom resource has not been created yet. Run the following command to create the custom resource instance
```bash
kubectl apply -f config/samples/fluxsecret_v1beta1_stackoverdrivesecret.yaml
```
With the flux operator deployed in production the flux operator will take this yaml definition and apply it to the cluster. So in essense the shortcomings of the flux operator will be fulfilled by this new custom operator. 

You will notice the secret created either in the logs or in the dashboard under the Secrets section. You can verify that
the secret is the same as the original private key by clicking on the eye icon. Or from the command line you can
```kubectl get secrets SECRET_NAME  -o yaml  --namespace flux-secrets-system```. Take the data and base64 decode it.
That should be the original private key.