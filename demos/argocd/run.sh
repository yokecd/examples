set -eux -o pipefail

# Make sure to have a recent installation of yoke
go install github.com/yokecd/yoke/cmd/yoke@latest

CLUSTER_NAME="demo-argocd"

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

yoke apply \
  -create-namespace \
  -namespace argocd \
  -wait 10m \
  argocd oci://ghcr.io/yokecd/yokecd-installer

kubectl -n argocd port-forward deployment/argocd-server 8080:8080 &

password=$(kubectl -n argocd get secrets argocd-initial-admin-secret --template '{{ .data.password }}' | base64 -d)

kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cats
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default

  source:
    repoURL: https://github.com/yokecd/examples.git
    targetRevision: main
    path: .

    plugin:
      name: yokecd
      parameters:
        - name: wasm
          string: https://github.com/yokecd/examples/releases/download/latest/demos_ingress_atc.wasm.gz
        - name: input
          string: |
            metadata:
              name: example
            spec:
              image: davidmdm/c4ts:latest
              replicas: 3
              pathPrefix: /miaow
              env: 
                PORT: '80'

  destination:
    server: https://kubernetes.default.svc
    namespace: default

  syncPolicy:
    automated: {}
EOF

cat <<EOF

To visit the ArgoCD dashboard open: http://localhost:8080

Login info:
username: admin
password: $password
EOF
