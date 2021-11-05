module github.com/mildred/terraform-provider-sys

require (
	github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7 // indirect
	github.com/coreos/go-systemd/v22 v22.3.1
	github.com/hashicorp/go-getter/v2 v2.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/terraform-plugin-docs v0.5.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.2
	github.com/hashicorp/terraform-plugin-test v1.4.3 // indirect
	github.com/otiai10/copy v1.5.1
)

go 1.13

replace github.com/coreos/go-systemd/v22 => github.com/vaspahomov/go-systemd/v22 v22.1.1-0.20201215170244-db69fcca5b95
