# Helm Inject Plugin

This is a Helm plugin which provides the ability to inject additional configuration during Helm release upgrade. It works like
`helm upgrade`, but with a `--inject` flag.

The default injector is [linkerd](https://linkerd.io/), but you can specify `--injector` to use any other executable in your $PATH with an `inject` command such as:
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
upgrade a release including inject

Usage:
  inject upgrade [RELEASE] [CHART] [flags]

Flags:
      --command string         injection command to be used (default "inject")
      --debug                  enable verbose output
      --dry-run                simulate an upgrade
  -h, --help                   help for upgrade
      --inject-flags strings   flags to be passed to injector, without leading "--" (can specify multiple). Example: "--inject-flags tls=optional,skip-inbound-ports=25,skip-inbound-ports=26"
      --injector string        injector to use (must be pre-installed) (default "linkerd")
  -i, --install                if a release by this name doesn't already exist, run an install
      --kubecontext string     name of the kubeconfig context to use
      --namespace string       namespace to install the release into (only used if --install is set). Defaults to the current kube config namespace
      --set stringArray        set values on the command line (can specify multiple)
      --timeout int            time in seconds to wait for any individual Kubernetes operation (like Jobs for hooks) (default 300)
      --tls                    enable TLS for request
      --tls-cert string        path to TLS certificate file (default: $HELM_HOME/cert.pem)
      --tls-key string         path to TLS key file (default: $HELM_HOME/key.pem)
  -f, --values stringArray     specify values in a YAML file or a URL (can specify multiple)
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
* If you want to pass any flags to the injector - use the `--inject-flags` flag.
* If you are using the `--kube-context` flag, you need to change it to `--kubecontext`, since helm plugins [drop this flag](https://github.com/helm/helm/blob/master/docs/plugins.md#a-note-on-flag-parsing).

### Examples

Check out the first example of a custom executable in the [examples](/examples) section!