set -x

CLUSTER=demo-dynamic-mode

kind delete cluster --name=$CLUSTER && kind create cluster --name=$CLUSTER

GOOS=wasip1 GOARCH=wasm go build -o ./demos/dynamic-mode/build-artifacts/flight.wasm ./demos/dynamic-mode/backend/flight
GOOS=wasip1 GOARCH=wasm go build -o ./demos/dynamic-mode/build-artifacts/setup.wasm ./demos/dynamic-mode/setup

yoke takeoff --debug --wait 5m demo ./demos/dynamic-mode/build-artifacts/setup.wasm
yoke takeoff --debug --wait 2m --namespace atc --create-namespace atc oci://ghcr.io/yokecd/atc-installer:latest

export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=root

kubectl port-forward svc/demo-vault 8200:8200 &

sleep 1
vault kv put secret/demo hello=world

go run ./demos/dynamic-mode/airway | yoke takeoff -debug -wait 1m demo-airway

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
