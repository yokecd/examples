set -x

CLUSTER=demo-dynamic-mode

# Create a temporary demo cluster
kind delete cluster --name=$CLUSTER && kind create cluster --name=$CLUSTER

# Build and execute our setup -- install vault and external-secrets-operator.
GOOS=wasip1 GOARCH=wasm go build -o ./demos/dynamic-mode/build-artifacts/setup.wasm ./demos/dynamic-mode/setup
yoke takeoff --debug --wait 5m demo ./demos/dynamic-mode/build-artifacts/setup.wasm

# Install the AirTrafficController using its latest OCI image.
yoke takeoff --debug --wait 2m --namespace atc --create-namespace atc oci://ghcr.io/yokecd/atc-installer:latest

# Open a port-forward to vault and create a a secret for our demo with hello=world Key-Value.
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

kubectl port-forward svc/demo-vault 8200:8200 &

sleep 1
vault kv put secret/demo hello=world

# Create our airway. We could have compiled it to wasm first, but given that the flight was completely static
# and did not depend on the release name or env, it is the same as just executing it and piping it to yoke.
# All roads lead to Rome.
go run ./demos/demo-dynamic-modede/airway | yoke takeoff -debug -wait 1m demo-airway

# Create a Backend corresponding to the airway we just created.
# Notice that we are mapping a secret into our deployment --
# Environment variable DEMO will the key hello found at secret/demo in vault.
kubectl apply -f - <<EOF
apiVersion: examples.com/v1
kind: Backend
metadata:
  name: demo-backend
spec:
  image: nginx:latest
  replicas: 2
  secrets:
    DEMO:
      path: secret/demo
      key: hello
EOF

# Run this commented command at your leisure to update the deployment by changing the secret in vault.

# VAULT_ADDR=http://localhost:8200 VAULT_TOKEN=root vault kv put secret/demo hello=fromtheotherside
