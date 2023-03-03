# LaSh (Landscape Shell)

LaSh is an open source CLI that can be used to install the PortalNOW Landscape IDP and Control planes.

The CLI allows bootstrapping any kubernetes cluster with the minimum required components to run Landscape IDP.


# Requirements

- A Kubernetes cluster with default storage-class, ingress controller and, the relative `kubeconfig` file;

For testing purposes a [Kind Cluster](kind-cluster.yaml) can be created using the [make](./Makefile) commands in this repo.

For example running `make kind-up` will create a Kind cluster with the name `local-dev` and the `kubeconfig` file will be stored in `~/.kube/config`.

# Installation

Follow the instructions in the [installation](./docs/INSTALL.md) guide to install LaSh.



