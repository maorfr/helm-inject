# Helm Inject Plugin

This is a Helm plugin which provides the ability to inject additional configuration during Helm release upgrade. It works like
`helm upgrade`, but with a `--inject` flag.

The default injector is [linkerd](https://linkerd.io/), but you can specify `--injector` to use any other binary with an `inject` command such as:
```
myInjector inject /path/to/file.yaml
```

## Usage

Inject linkerd proxy sidecar during Helm upgrade

```
$ helm inject upgrade [flags]
```

### Flags:

```
$ helm inject upgrade --help
upgrade a release including inject (default injector: linkerd)

Usage:
  inject upgrade [RELEASE] [CHART] [flags]

Flags:
      --debug                 enable verbose output
      --dry-run               simulate an upgrade
  -h, --help                  help for upgrade
      --injector string       injector to use (must be pre-installed) (default "linkerd")
  -i, --install               if a release by this name doesn't already exist, run an install
      --kube-context string   name of the kubeconfig context to use
      --namespace string      namespace to install the release into (only used if --install is set). Defaults to the current kube config namespace
      --set stringArray       set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --tls                   enable TLS for request
      --tls-cert string       path to TLS certificate file (default: $HELM_HOME/cert.pem)
      --tls-key string        path to TLS key file (default: $HELM_HOME/key.pem)
  -f, --values stringArray    specify values in a YAML file or a URL(can specify multiple)
```


## Install

```
$ helm plugin install https://github.com/maorfr/helm-inject
```

The above will fetch the latest binary release of `helm inject` and install it.

### Developer (From Source) Install

If you would like to handle the build yourself, instead of fetching a binary, this is how recommend doing it.

First, set up your environment:

- You need to have [Go](http://golang.org) installed. Make sure to set `$GOPATH`
- If you don't have [Glide](http://glide.sh) installed, this will install it into
  `$GOPATH/bin` for you.

Clone this repo into your `$GOPATH`. You can use `go get -d github.com/maorfr/helm-inject`
for that.

```
$ cd $GOPATH/src/github.com/maorfr/helm-inject
$ make bootstrap build
$ SKIP_BIN_INSTALL=1 helm plugin install $GOPATH/src/github.com/maorfr/helm-inject
```

That last command will skip fetching the binary install and use the one you built.

### Notes

* Not all `helm upgrade` flags are added. If you need any other flags from `helm upgrade` - you are welcome to open an issue, or even submit a PR.
* Inject currently does not take flags. If you need any flags for the default injector (linkerd) - you are welcome to open an issue, or even submit a PR.
