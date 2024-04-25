# Running in docker  
  
Run Docker image from `gchr.io/patrickdappollonio/kubectl-slice`

## Usage
The container is build to execute the the kubectl-slice tool inline. 
   
Example:  
`docker run --rm -v "${PWD}/slice/testdata":workdir kubectl-slice -f ingress-namespace.yaml -o ./`  
  
This will split the `ingress-namespace.yaml` file in the directory `${PWD}/slice/testdata`.  
  
For help use:  
`docker run --rm -v "${PWD}/slice/testdata":workdir kubectl-slice -h`

Statements to `kubectl-slice` can be wrapped in `""` if the commandline escapes unintentionally.  
Example:  
`docker run --rm -v "${PWD}/slice/testdata":workdir kubectl-slice "-h"`
  
## Manual build the docker image  
Requires docker or another OCI tool.  

Follow these steps:  
1. Clone the repo and cd into the project
1. Run `docker build . -t kubectl-slice`

## Manual push the docker image  
A 3rd. party container registry like docker hub is needed to do this.  

1. Run `docker tag kubectl-slice:latest YOUR_REGISTRY.com/ORG/kubectl-slice:1.0`
1. Run `docker push YOUR_REGISTRY.com/ORG/kubectl-slice:1.0`

## Running arm locally but x86/x64 on the endpoint

Requires buildkit from (moby)[https://github.com/moby/buildkit] installed.  

Example change the dockerfile:  
`FROM --platform=linux/amd64 golang:1.20.3 as builder`  
`FROM --platform=linux/amd64 alpine:3.14 as production`  
  
Example run direct inline build command instead:  
`Docker buildx build --platform linux/amd64 . -t kubectl-slice`