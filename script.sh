# install yoke cli
go install github.com/yokecd/yoke/cmd/yoke@latest

# create a local cluster
kind delete cluster && kind create cluster

# install the atc
yoke takeoff -wait 30s --namespace atc atc 'https://github.com/yokecd/yoke/releases/download/atc-installer%2Fv0.3.0/atc-installer.wasm.gz'

# install the yokcd/examples Backend-Airway
yoke takeoff -wait 30s backendairway "https://github.com/yokecd/examples/releases/download/latest/atc_backend_airway.wasm.gz"

# You are done! You can now create Backends!
kubectl apply -f - <<EOF
apiVersion: examples.com/v1
kind: Backend
metadata:
  name: nginx
spec:
  image: nginx:latest
  replicas: 2
EOF
