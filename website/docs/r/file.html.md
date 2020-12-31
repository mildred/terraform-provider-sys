---
layout: "remote"
page_title: "Remote: remote_file"
sidebar_current: "docs-remote-resource-file"
description: |-
  Generates a remote file from content.
---

# remote_file

Generates a remote file with the given content.

~> **Note** When working with remote files, Terraform will detect the resource
as having been deleted each time a configuration is applied on a new machine
where the file is not present and will generate a diff to re-create it. This
may cause "noise" in diffs in environments where configurations are routinely
applied by many different users or within automation systems.

## Example Usage

```hcl
resource "remote_file" "foo" {
    content     = "foo!"
    filename = "${path.module}/foo.bar"
}
```

## Argument Reference

The following arguments are supported:

* `source` - (Optional) The source file to copy, compatible with go-getter.

* `content` - (Optional) The content of file to create. Conflicts with `sensitive_content` and `content_base64`.

* `sensitive_content` - (Optional) The content of file to create. Will not be displayed in diffs. Conflicts with `content` and `content_base64`.

* `content_base64` - (Optional) The base64 encoded content of the file to create. Use this when dealing with binary data. Conflicts with `content` and `sensitive_content`.

* `filename` - (Required or `target_directory` must be present) The path of the file to create.

* `target_directory` - (Required or `filename` must be present) The path of target directory where the file should be put, must not exists unless `force_overwrite` is `true`. Upon resource deletion, the target directory will be entorely removed with no additional check. Can be useful when the source is an archive that go-getter extracts (it will refuse to do so with `filename`).

* `file_permission` - (Optional) The permission to set for the created file. Expects an a string. The default value is `"0777"`.

* `directory_permission` - (Optional) The permission to set for any directories created. Expects a string. The default value is `"0777"`.

* `force_overwrite` - (Optional, default: `false`) When `true`, allows to overwrite target file or directory.

Any required parent directories will be created automatically, and any existing file with the given name will be overwritten.
