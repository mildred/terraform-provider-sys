package sys

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/hashicorp/go-getter/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mildred/terraform-provider-sys/sys/file_getter"
	"github.com/mildred/terraform-provider-sys/sys/utils"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFileCreate,
		ReadContext:   resourceFileRead,
		DeleteContext: resourceFileDelete,
		UpdateContext: resourceFileUpdate,

		Description: `
sys_file generates a local, similarly to local_file, with a number of options. Files or directories can be generated from:
- direct file content (plain, base64 or sensitive)
- source file to copy, with the ability to fetch remote repositories or files from archives using go-getter.
- source directory to copy (local or remote using go-getter).

Any required parent directories will be created automatically, and any existing file with the given name will be overwritten.

If the destination file exists, creation will block. However the resource has the ability to remove it if it exists prior to running in every case, or force overwriting it. if the source is a local file or directory, it can generate a symlink too (default behaviour of go-getter).
`,

		Schema: map[string]*schema.Schema{
			"content": {
				Description:   "The content of file to create. Conflicts with `sensitive_content` and `content_base64`.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content_base64", "source"},
			},
			"sensitive_content": {
				Description:   "The content of file to create. Will not be displayed in diffs. Conflicts with `content` and `content_base64`.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"content", "content_base64", "source"},
			},
			"content_base64": {
				Description:   "The base64 encoded content of the file to create. Use this when dealing with binary data. Conflicts with `content` and `sensitive_content`.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sensitive_content", "content", "source"},
			},
			"source": {
				Description:   "The source file to copy, compatible with go-getter.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"content", "sensitive_content", "content_base64"},
			},
			"filename": {
				Description:   "(Required unless `target_directory` is specified) The path of the file to create.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"target_directory"},
			},
			"target_directory": {
				Description:   "(Conflicts with `filename` or `content*`) The path of target directory where the file should be put, must not exists unless `force_overwrite` is `true`. Upon resource deletion, the target directory will be entorely removed with no additional check. Can be useful when the source is an archive that go-getter extracts (it will refuse to do so with `filename`).",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filename", "content", "sensitive_content", "content_base64"},
			},
			"file_permission": {
				Description:  "(default: \"0666\") The permission to set for the created file. Expects an a string.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "0666",
				ValidateFunc: validateMode,
			},
			"directory_permission": {
				Description:  "(default: \"0777\") The permission to set for any directories created. Expects a string.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "0777",
				ValidateFunc: validateMode,
			},
			"force_overwrite": {
				Description: "(default: false) When `true`, allows to overwrite target file or directory.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"clear_destination": {
				Description: "(default: false) Remove directory destination before recreating it. Must be used with force_overwrite",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"symlink_destination": {
				Description: "(default: false) Symlink destination if source is a directory and target_directory is set.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"unlink_before_create": {
			        Description: "Unlink file before creating it (allows to use a new inode)",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

type resourceFileSystemd struct {
	unit       string
	enable     bool
	has_enable bool
	start      bool
	has_start  bool
	old_enable bool
	had_enable bool
	old_start  bool
	had_start  bool
}

func resourceFileRead(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	outputPath, isDir, err := getDestination(d)
	if err != nil {
		return diag.Errorf("cannot get destination, %v", err)
	}

	// If the output file doesn't exist, mark the resource for creation.
	st, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		d.SetId("")
		return nil
	}

	same, err := utils.FileModeSame(d.Get("file_permission").(string), st.Mode(), utils.Umask)
	if err != nil {
		return diag.Errorf("checking file mode, %v", err)
	}
	if !same {
		d.Set("file_permission", st.Mode().String())
	}

	// Verify that the content of the destination file matches the content we
	// expect. Otherwise, the file might have been modified externally and we
	// must reconcile.
	if !isDir {
		outputContent, err := ioutil.ReadFile(outputPath)
		if err != nil {
			return diag.Errorf("cannot read file, %v", err)
		}

		outputChecksum := sha1.Sum([]byte(outputContent))
		if hex.EncodeToString(outputChecksum[:]) != d.Id() {
			d.SetId("")
			return nil
		}
	} else {
		sum, err := checksumFile(outputPath)
		if err != nil {
			return diag.Errorf("cannot checksum %s, %v", outputPath, err)
		}
		d.SetId(sum)
	}

	return nil
}

func resourceFileContent(d *schema.ResourceData) ([]byte, bool, error) {
	if content, sensitiveSpecified := d.GetOk("sensitive_content"); sensitiveSpecified {
		return []byte(content.(string)), true, nil
	}
	if b64Content, b64Specified := d.GetOk("content_base64"); b64Specified {
		res, err := base64.StdEncoding.DecodeString(b64Content.(string))
		return res, true, err
	}
	if content, contentSpecified := d.GetOk("content"); contentSpecified {
		return []byte(content.(string)), true, nil
	}
	return nil, false, nil
}

func resourceFileUpdate(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	destination, is_directory, err := getDestination(d)
	if err != nil {
		return diag.Errorf("destination, %s", err)
	}

	var perm_name string
	if is_directory {
		perm_name = "directory_permission"
	} else {
		perm_name = "file_permission"
	}

	if d.HasChange(perm_name) {
		perm := d.Get(perm_name).(string)
		modeInt, _ := strconv.ParseInt(perm, 8, 64)
		mode := os.FileMode(modeInt)

		err := os.Chmod(destination, mode)
		if err != nil {
			return diag.Errorf("cannot chmod %s, %s", mode, err)
		}
	}

	return nil
}

func getDestination(d *schema.ResourceData) (string, bool, error) {
	var destination = ""
	var is_directory bool
	var good bool

	if filename, ok := d.GetOk("filename"); ok {
		destination = filename.(string)
		is_directory = false
		good = true
	}
	if target_directory, ok := d.GetOk("target_directory"); ok {
		destination = target_directory.(string)
		is_directory = true
		good = true
	}

	if !good {
		return "", false, fmt.Errorf("missing filename or target_directory")
	}

	return destination, is_directory, nil
}

func readFileOrDir(w io.Writer, filename string, st os.FileInfo) error {
	var err error
	if st == nil {
		st, err = os.Lstat(filename)
		if err != nil {
			return fmt.Errorf("lstat %s, %e", filename, err)
		}
	}

	if st.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(filename)
		if err != nil {
			return fmt.Errorf("readlink %s, %v", filename, err)
		}

		fmt.Fprintf(w, "%s", link)
		return nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file %s, %v", filename, err)
	}
	defer f.Close()

	if st.IsDir() {
		files, err := f.Readdir(-1)
		if err != nil {
			return fmt.Errorf("readdir %s, %e", filename, err)
		}
		sort.Slice(files, func(a, b int) bool {
			return files[a].Name() < files[b].Name()
		})
		for _, fst := range files {
			fmt.Fprintf(w, "%d.%s.%d.", len(fst.Name()), fst.Name(), fst.Size())
			readFileOrDir(w, path.Join(filename, fst.Name()), fst)
		}
	} else if st.Mode().IsRegular() {
		_, err = io.Copy(w, f)
		if err != nil {
			return fmt.Errorf("reading %s, %v", filename, err)
		}
	} else {
		return fmt.Errorf("cannot handle %s type %s", filename, st.Mode().String())
	}
	return nil
}

func checksumFile(destination string) (string, error) {
	h := sha1.New()
	err := readFileOrDir(h, destination, nil)
	if err != nil {
		return "", err
	}
	checksum := h.Sum(nil)
	return hex.EncodeToString(checksum[:]), nil
}

func resourceFileCreate(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	forceOverwrite := d.Get("force_overwrite").(bool)
	clearDestination := d.Get("clear_destination").(bool)
	unlinkBeforeCreate := d.Get("unlink_before_create").(bool)
	symlink_destination := d.Get("symlink_destination").(bool)
	source, sourceSpecified := d.GetOk("source")
	content, contentSpecified, err := resourceFileContent(d)
	if err != nil {
		return diag.Errorf("content error, %v", err)
	}

	destination, is_directory, err := getDestination(d)
	if err != nil {
		return diag.Errorf("finding destination, %v", err)
	}

	dirPerm := d.Get("directory_permission").(string)
	dirMode, _ := strconv.ParseInt(dirPerm, 8, 64)

	destinationDir := path.Dir(destination)
	if _, err := os.Stat(destinationDir); err != nil {
		if err := os.MkdirAll(destinationDir, os.FileMode(dirMode)); err != nil {
			return diag.Errorf("cannot create parent directories, %v", err)
		}
	}

	filePerm := d.Get("file_permission").(string)
	fileMode, _ := strconv.ParseInt(filePerm, 8, 64)

	if sourceSpecified {
		if !forceOverwrite {
			if _, err := os.Lstat(destination); err == nil || !os.IsNotExist(err) {
				return diag.Errorf("destination exists at %v", destination)
			}
		}
		if forceOverwrite && clearDestination && is_directory {
			err := os.RemoveAll(destination)
			if err != nil {
				return diag.Errorf("cannot delete target directory, %v", err)
			}
		} else if unlinkBeforeCreate {
		        err := os.Remove(destination)
			if err != nil {
				return diag.Errorf("cannot unlink target before creation, %v", err)
			}
		}
		get := &getter.Client{
			Getters:       getter.Getters,
			Decompressors: getter.Decompressors,
		}

		if !symlink_destination {
			fileGetter := new(file_getter.FileGetter)
			get.Getters = append([]getter.Getter{fileGetter}, get.Getters...)
		}

		var mode = getter.ModeFile
		if is_directory {
			mode = getter.ModeAny
		}

		_, err = get.Get(ctx, &getter.Request{
			Src:     source.(string),
			Dst:     destination,
			GetMode: mode,
			Copy:    !symlink_destination,
		})

		if err != nil {
			return diag.Errorf("cannot fetch source %v, %v", source, err)
		}
	}

	if contentSpecified {
		data := []byte(content)
		flags := os.O_WRONLY | os.O_CREATE
		if forceOverwrite {
			flags = flags | os.O_EXCL
		} else {
			flags = flags | os.O_TRUNC
		}
		f, err := os.OpenFile(destination, flags, os.FileMode(fileMode))
		if err != nil {
			return diag.Errorf("cannot write file, %v", err)
		}
		n, err := f.Write(data)
		if err == nil && n < len(data) {
			err = io.ErrShortWrite
		}
		if err1 := f.Close(); err == nil {
			err = err1
		}
		if err != nil {
			return diag.Errorf("cannot write file, %v", err)
		}

		checksum := sha1.Sum([]byte(content))
		d.SetId(hex.EncodeToString(checksum[:]))
	} else {
		if is_directory {
			err = os.Chmod(destination, os.FileMode(dirMode))
		} else {
			err = os.Chmod(destination, os.FileMode(fileMode))
		}
		if err != nil {
			return diag.Errorf("cannot chmod %s, %v", filePerm, err)
		}
		id, err := checksumFile(destination)
		if err != nil {
			return diag.Errorf("cannot checksum file %s, %v", destination, err)
		}
		d.SetId(id)
	}

	return nil
}

func resourceFileDelete(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if filename := d.Get("filename").(string); filename != "" {
		err := os.Remove(filename)
		if err != nil {
			return diag.Errorf("cannot delete file, %v", err)
		}
	}

	if target_directory := d.Get("target_directory").(string); target_directory != "" {
		err := os.RemoveAll(target_directory)
		if err != nil {
			return diag.Errorf("cannot delete target directory, %v", err)
		}
	}

	return nil
}
