set -eux

# install the atc
yoke takeoff -debug -wait 1m --create-namespace --namespace atc atc 'https://github.com/yokecd/yoke/releases/download/latest/atc-installer.wasm.gz'

# install the yokcd/examples Backend-Airway
yoke takeoff -debug -wait 1m backendairway "https://github.com/yokecd/examples/releases/download/latest/atc_backend_airway.wasm.gz"

# You are done! You can now create Backends!
kubectl apply -f - <<EOF
apiVersion: examples.com/v1
kind: Backend
metadata:
  name: nginx
spec:
  image: nginx:latest
  replicas: 2
  labels:
    originalVersion: v1
EOF
