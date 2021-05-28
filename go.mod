module github.com/mildred/terraform-provider-sys

require (
	github.com/coreos/go-systemd/v22 v22.3.1
	github.com/hashicorp/go-getter/v2 v2.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/terraform-exec v0.12.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.2
	github.com/hashicorp/terraform-plugin-test v1.4.3 // indirect
	github.com/otiai10/copy v1.5.1
	github.com/zclconf/go-cty v1.7.1 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
)

go 1.13

replace github.com/coreos/go-systemd/v22 => github.com/vaspahomov/go-systemd/v22 v22.1.1-0.20201215170244-db69fcca5b95
