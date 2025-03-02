set -eux

# install yoke cli
go install github.com/yokecd/yoke/cmd/yoke@latest

# create a local cluster
kind delete cluster

kind create cluster --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
featureGates:
  InformerResourceVersion: true
  StorageVersionMigrator: true
  APIServerIdentity: true
runtimeConfig:
  'storagemigration.k8s.io/v1alpha1': true
EOF

# install the atc
yoke takeoff -wait 30s --create-namespace --namespace atc atc 'https://github.com/yokecd/yoke/releases/download/latest/atc-installer.wasm.gz'

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
  labels:
    originalVersion: v1
EOF

# update the backendairway to use the v2 version.
yoke takeoff -wait 30s backendairway "https://github.com/yokecd/examples/releases/download/latest/atc_backend_airway_v2.wasm.gz"

# Now we can create v2 versions
kubectl apply -f - <<EOF
apiVersion: examples.com/v2
kind: Backend
metadata:
  name: nginx-v2
spec:
  img: nginx:latest
  replicas: 2
  meta:
    labels:
      originalVersion: v1
    annotations:
      cool: 'yes'
EOF
