package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mildred/terraform-provider-remote/sys"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sys.Provider})
}
