package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mildred/terraform-provider-sys/sys"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sys.Provider})
}
