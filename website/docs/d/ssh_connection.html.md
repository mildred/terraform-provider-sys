---
layout: "remote"
page_title: "Remote: remote_ssh_connection"
sidebar_current: "docs-remote-datasource-ssh_connection"
description: |-
  Establish a remote connection
---

# remote_ssh_connection

`remote_ssh_connection` establish a remote connection.

## Example Usage

```hcl
data "remote_ssh_connection" "host" {
    host = "hostname.example.org"
    user = "admin"
    sudo = true
}
```

## Argument Reference

The following argument is required:

* `host` - (Required) The hostname to connect to.
* `user` - (Optional) The user to use for connection.
* `sudo` - (Optional) If `sudo` should be used on the host.
* `port` - (Optional) SSH port number to use.

## Attributes Exported

The following attribute is exported:

* `conn` - The connection string used by all other resources.
