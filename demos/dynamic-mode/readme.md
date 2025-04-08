# Dynamic-Mode

This demo illustrates the power of dynamic-mode airways.

Dynamic-mode airways enable your packages to respond to changes in their subresources dynamically.
Instead of being constrained to a single desired state, they allow the desired state of your package to update as side effects occur in the real world.

This demo addresses an age-old Kubernetes question using dynamic airways:

_How do we restart a deployment when a secret changes?_

In theory, the demo could have simply involved an Airway that describes a custom resource (e.g., Backend) to deploy a secret and a deployment, then update the deployment whenever the secret changes. However, that approach might feel too direct and not representative of a real-world setup. Instead, this demo sets up an in-cluster development HashiCorp Vault and uses the external-secrets-operator to manage secrets. This, in turn, triggers updates to the deployment whenever the secrets change.

As a result, the deployment is redeployed every time the secrets in Vault are updated.

Let's get started!

## Prerequisites

- Go 1.24+
- Vault CLI
- Yoke CLI (latest recommended)
- kind CLI (Kubernetes in Docker)

## Run the Demo

From the root of this repository, run:

```bash
./demos/dynamic-mode/run.sh
```

This script will perform the following steps:

1. Kill and restart a Kind Cluster named "demo-dynamic-mode."
2. Build a setup WebAssembly (Wasm) that embeds the Vault and external-secrets-operator charts.
3. Run the setup as a Yoke release.
4. Install the Air Traffic Controller using the latest version.
5. Create a port-forward to the Vault running in your cluster from the setup.
6. Add a secret to Vault.
7. Create an Airway for the example Backend type.
8. Create a Backend instance that uses the secret defined in Vault.

Afterward, you can inspect your cluster to confirm that the deployment exists with a secret environment variable.

If you update the secret in Vault, the deployment will be redeployed!

For example:

```bash
VAULT_TOKEN=root VAULT_ADDR=http://localhost:8200 vault kv put secret/demo hello=goodbye
```

