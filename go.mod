module github.com/mildred/terraform-provider-sys

require (
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/andybalholm/crlf v0.0.0-20171020200849-670099aa064f // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.1
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/go-getter v1.5.3 // indirect
	github.com/hashicorp/go-getter/v2 v2.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/terraform-plugin-docs v0.18.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.0-rc.2
	github.com/hashicorp/terraform-plugin-test v1.4.3 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/otiai10/copy v1.5.1
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
)

go 1.13

replace github.com/coreos/go-systemd/v22 => github.com/vaspahomov/go-systemd/v22 v22.1.1-0.20201215170244-db69fcca5b95
