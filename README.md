# Helm Inject Plugin

This is a Helm plugin which provides the ability to inject additional configuration during Helm release upgrade. It works like
`helm upgrade`, but with a `--inject` flag.

The default injector is [linkerd](https://linkerd.io/), but you can specify `--injector` to use any other binary with an `inject` command:
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
