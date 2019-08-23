---
layout: "remote"
page_title: "Remote: remote_file"
sidebar_current: "docs-remote-datasource-file"
description: |-
  Reads a file from the remote filesystem.
---

# remote_file

`remote_file` reads a file from the remote filesystem.

## Example Usage

```hcl
data "remote_file" "foo" {
    filename = "${path.module}/foo.bar"
}
```

## Argument Reference

The following argument is required:

* `conn` - (Required) The connection string.
* `filename` - (Required) The path to the file that will be read. The data
  source will return an error if the file does not exist.

## Attributes Exported

The following attribute is exported:

* `content` - The raw content of the file that was read.
* `content_base64` - The base64 encoded version of the file content (use this when dealing with binary data).

The content of the file must be valid UTF-8 due to Terraform's assumptions
about string encoding. Files that do not contain UTF-8 text will have invalid
UTF-8 sequences replaced with the Unicode replacement character.
