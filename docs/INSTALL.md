# LaSh

Command line tool to manage PlatformNOW Landscape IDP.

## Requirements

- A Kubernetes cluster with default storage-class, ingress controller and, the relative `kubeconfig` file;

# Installation

## Download and Install LaSh

### MacOs

```sh
brew tap platformnow/lash
brew install lash
```

or if you have already installed LaSh using brew, you can upgrade LaSh by running:

```sh
brew upgrade lash
```

### From the [Binary Releases](https://github.com/platfornow/lash/releases) (macOS, Windows, Linux)

LaSh currently provides pre-built binaries for the following:

- macOS (Darwin)
- Windows
- Linux

1. Download the appropriate version for your platform from [Lash Releases](https://github.com/platfornow/lash/releases).

2. Once downloaded unpack the archive (zip for Windows; tarball for Linux and macOS) to extract the executable binary. 

3. If you want to use from any location you must put the binary executable in your `Path` or add the directory where is it to the environment variables.

### Using the [`Go`](https://go.dev/dl/) toolchain

```sh
go install github.com/platfornow/lash@latest
```

## Install LaSh

```sh
lash init
```

### Syntax

Most of the commands have flags; you can specify these:

- using the short notation (single dash and single letter; i.e.: `-v`)
- using the long notation (double dash and full flag name; i.e.: `--verbose`)
- by specifying environment variables
- from a config file located in `$HOME/.lash/lash.yaml`

## Initialize LaSh Platform

Usage: **`LaSh init [flags]`** where:

| Flag                       | Description                                                          | Default                                    |
|:---------------------------|:---------------------------------------------------------------------|:-------------------------------------------|
| `--catalog-url`            | control plane url                                                    | https://github.com/platformnow/catalog.git |
| `--context`                | kube context                                                         | current context                            |
| `--help`                   | help for init                                                        | n/a                                        |
| `--http-proxy`             | use the specified HTTP proxy                                         | value of `HTTP_PROXY` env var              |
| `--https-proxy`            | use the specified HTTPS proxy                                        | value of `HTTPS_PROXY` env var             |
| `--no-proxy`               | comma-separated list of hosts and domains which do not use the proxy | value of `NO_PROXY` env var                |
| `-m, --management-cluster` | create a management cluster fro this cluster                         | false                                      |
| `-n, --namespace`          | namespace where to install landscape                                 | landscape-system                           |
| `--no-crossplane`          | dont install crossplane                                              | false                                      |
| `-v, --verbose`            | print verbose output                                                 | false                                      |

Example:

```sh
lash init
```

# Uninstall

```sh
lash uninstall
```
