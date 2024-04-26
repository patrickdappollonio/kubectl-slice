# Running in docker  
  
Run the Docker image from `gchr.io/patrickdappollonio/kubectl-slice`

## Usage
The container is build to execute both interactive and inline.  

### Interactive shell
  
Example, this will start the container in the `/workdir` directory.
```sh
docker run --rm -it \
gchr.io/patrickdappollonio/kubectl-slice /bin/sh
```
  
### Inline
   
**Example 1:**  
This will split the `ingress-namespace.yaml` file in the directory `${PWD}/slice/testdata`.  

```sh
docker run --rm -v \
"${PWD}/slice/testdata":/workdir gchr.io/patrickdappollonio/kubectl-slice \
kubectl-slice -f ingress-namespace.yaml -o ./
```
  
**Example 2:**  
To display the `kubectl-slice` help.  
    
```sh
docker run --rm -v \
"${PWD}/slice/testdata":/workdir gchr.io/patrickdappollonio/kubectl-slice \
kubectl-slice -h
```

**Example 3:**  
Statements to `kubectl-slice` can be wrapped in `""` if the commandline escapes unintentionally.   
  
```sh
docker run --rm -v \
"${PWD}/slice/testdata":/workdir gchr.io/patrickdappollonio/kubectl-slice \
kubectl-slice "-h"
```