# Dynamic-Mode

This demo is meant to illustrate the power of dynamic mode airways.

Dynamic mode airways allow your packages to react to changes to their submodules.
In this way we are not bound to a single desired state, but can update the desired state of our package
as side effects happen in the real world.

This demo will take an age old question in Kubernetes and solve it using dynamic airways:

_How do we restart a deployment when a secret changes?_

The demo could have been an Airway that describes some Custom Type (Backend) that simply deploys a secret
and deployment and updates deployment whenever the secret changes. But that would have been to direct,
and perhaps does not feel like a real-world setup. Hence this demo will setup a in-cluster development hashicorp vault
and use the external-secrets-operator to update our secrets which in turn updates our deployment.

Therefore, we redeploy our deployment everytime we update our secrets in vault.

Let's get started!

## Prerequisites

- Go 1.24+
- Vault CLI
- Yoke CLI (latest recommended)
- kind CLI (Kubernetes in Docker)

## Run the demo

From the root of this repository run:

```bash
./demos/dynamic-mode/run.sh
```

It will do the following:

- kill and restart a Kind Cluster named "demo-dynamic-mode"
- Build a setup wasm that embeds the vault and external-secrets-operator charts.
- Run the setup as a yoke release.
- Install the Air Traffic Controller using the latest version.
- Create a port-forward to vault running in your cluster from the setup
- Add a secret to vault.
- Create an Airway for our example Backend type.
- Create a Backend that uses the secret we defined in vault.

You can then inspect your cluster to notice that the deployment exists with a secret env var.

If you update the vault secret the deployment redeploys!

```bash
VAULT_TOKEN=root VAULT_ADDR=http://localhost:8200 vault kv put secret/demo hello=goodbye
```
