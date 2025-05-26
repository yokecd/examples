set -eux -o pipefail

# Make sure to have a recent installation of yoke
go install github.com/yokecd/yoke/cmd/yoke@latest

CLUSTER_NAME="demo-ingress"

# Delete and recreate a kind cluster called demo-ingress.
# This cluster contains a host-port mapping so that we can send requests to the cluster over localhost.
kind delete cluster --name=$CLUSTER_NAME && kind create cluster --name=$CLUSTER_NAME --config=- <<EOF
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  nodes:
  - role: control-plane
    extraPortMappings:
    - containerPort: 80
      hostPort: 80
      protocol: TCP
    - containerPort: 443
      hostPort: 443
      protocol: TCP
EOF

# Get the nginx ingress deployment from kind.sigs.k8s.io examples.
# This will create an nginx ingress controller listening for requests on our hostports 80 & 443,
# allowing us to send traffic and route it to our services via ingress.
#
# We pipe the yaml into yoke, so that a yoke release is created.
# We use the -namespace flag match the namespaces that are hard-coded in the yaml.
#
# We wait a maximum of 5m (realistically it should be much sooner) for any workloads to complete or become ready.
#
# The final argument is the release name: ingress-nginx
curl https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml |
  yoke apply \
    -namespace ingress-nginx \
    -debug \
    -wait 5m \
    ingress-nginx

cat <<EEOF

---

# You are setup with a new Kind cluster and ingress-nginx controller.
# To deploy a service run the next few lines!

# Create our wasm module using our flight implementation.
GOOS=wasip1 GOARCH=wasm go build -o ./demo.wasm ./demos/ingress/flight

yoke apply -wait 2m foo ./demo.wasm <<EOF
  image: ealen/echo-server:latest
  pathPrefix: /foo
  env:
    ENABLE__REQUEST: false
    ENABLE__ENVIRONMENT: false
EOF
EEOF
