set -x

CLUSTER=demo-dynamic-mode

kind delete cluster --name=$CLUSTER && kind create cluster --name=$CLUSTER

export GOOS=wasip1
export GOARCH=wasm

go build -o ./demos/dynamic-mode/build-artifacts/flight.wasm ./demos/dynamic-mode/backend/flight
go build -o ./demos/dynamic-mode/build-artifacts/setup.wasm ./demos/dynamic-mode/setup
go build -o ./demos/dynamic-mode/build-artifacts/airway.wasm ./demos/dynamic-mode/airway

yoke takeoff --debug --wait 5m demo ./demos/dynamic-mode/build-artifacts/setup.wasm
yoke takeoff --debug --wait 2m --namespace atc --create-namespace atc oci://ghcr.io/yokecd/atc-installer:latest

kubectl port-forward svc/demo-vault 8200:8200 &
