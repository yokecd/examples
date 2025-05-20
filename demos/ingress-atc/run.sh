set -eux -o pipefail

# Make sure to have a recent installation of yoke
go install github.com/yokecd/yoke/cmd/yoke@latest

# Delete and recreate a kind cluster called demo-ingress.
# This cluster contains a host-port mapping so that we can send requests to the cluster over localhost.
kind delete cluster --name=demo-ingress-atc && kind create cluster --name=demo-ingress-atc --config=- <<EOF
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

# Install the Air Traffic Controller.
yoke apply \
  -create-namespace \
  -namespace atc \
  -debug \
  -wait 5m \
  atc oci://ghcr.io/yokecd/atc-installer:latest

cat <<EEOF

---

# Install the Airway
go run ./demos/ingress-atc/backend/airway | yoke apply -wait 5m -debug backend-airway

# Create Backends regularly with kubectl
kubectl apply -f - <<EOF
  apiVersion: examples.com/v1
  kind: Backend
  metadata:
    name: echo
  spec:
    image: ealen/echo-server:latest
    pathPrefix: /echo
    env:
      ENABLE__REQUEST: false
      ENABLE__ENVIRONMENT: false
EOF
EEOF
