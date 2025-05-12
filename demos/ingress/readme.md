# Demo Ingress

This demo serves as an example of a simple deployment.

We build a flight which outputs a Deployment, Service, and Ingress.

To make this demo easy to run, the setup uses an ephemeral `kind` cluster exposing a hostPort on port 80.
This allows us to make requests to our cluster over localhost.

An nginx-ingress controller is installed using the example that can be found at:
https://kind.sigs.k8s.io/docs/user/ingress

This simulates locally the real world experience of deploying services into kubernetes and exposing them to the world.

It is meant to give the user a feeling of what defining a yoke Flight in Go feels like, as well as interacting with the Yoke CLI.

## Running the demo

To run the demo, execute the `run.sh` script found in the directory of the demo.

Read the instructions printed via the script to build and execute the Flight.

Feel free to play around with the flight implementation and see how powerful yoke can be for yourself!

## Requisite Dependencies

- Go Toolchain
- yoke
- kind
