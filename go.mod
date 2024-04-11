module github.com/mildred/terraform-provider-sys

require (
	github.com/coreos/go-systemd/v22 v22.3.1
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-getter v1.5.3 // indirect
	github.com/hashicorp/go-getter/v2 v2.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.2
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/otiai10/copy v1.5.1
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/zclconf/go-cty v1.14.1 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

go 1.13

replace github.com/coreos/go-systemd/v22 => github.com/vaspahomov/go-systemd/v22 v22.1.1-0.20201215170244-db69fcca5b95
