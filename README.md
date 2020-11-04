Terraform Provider
==================

The sys terraform provider allows provisionning configuration management
resource to local hosts. This is the equivalent of the `local` terraform
provider, but with more features, including script resources.

The idea was that traditional configuration management tools were lacking
important features that terraform has. In particular the tfstate is a great
feature that allows to rollback the deployment to its initial state. With
configuration management tools, the only source of truth is your code and when
you modify it, some resources created from previous executions might be left
over, creating some unclean state.


Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Building The Provider
---------------------

Run `make build` outside GOPATH or with `GO111MODULE=on`

Using the provider
------------------

Place the `terraform-provider-remote` executable in `~/.terraform.d/plugins`

You can do so with `make user-install`

See: https://www.terraform.io/docs/configuration/providers.html#third-party-plugins

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-remote
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

Making a release
----------------

From the [upstream documentation](https://www.terraform.io/docs/registry/providers/publishing.html):

- `export GPG_FINGERPRINT=01230FD4CC29DE17`
- `export GITHUB_TOKEN=...`
- Cache passphrase with `gpg --armor --detach-sign --local-user $GPG_FINGERPRINT </dev/null`
- Create tag: `git tag -s -u $GPG_FINGERPRINT vx.x.x`
- Make release: `goreleaser release --rm-dist`

