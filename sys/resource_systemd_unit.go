package sys

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	systemd "github.com/coreos/go-systemd/v22/dbus"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceSystemdUnit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemdUnitCreate,
		ReadContext:   resourceSystemdUnitRead,
		DeleteContext: resourceSystemdUnitDelete,
		UpdateContext: resourceSystemdUnitUpdate,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enable": {
				Type:        schema.TypeBool,
				Description: "Enable the unit",
				Optional:    true,
			},
			"mask": {
				Description: "Mask the unit",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"start": {
				Description: "Start the unit",
				Optional:    true,
			},
			"restart_on": {
				Type:        schema.TypeMap,
				Description: "Restart unit if this changes",
				Optional:    true,
			},
			"rollback": {
				Type:        schema.TypeMap,
				Description: "Rollback information to restore once the unit is destroyed",
				Computed:    true,
			},
			"system": {
				Type:          schema.TypeBool,
				Description:   "System systemd socket",
				Optional:      true,
				ConflictsWith: []string{"user"},
			},
			"user": {
				Type:          schema.TypeBool,
				Description:   "User systemd socket",
				Optional:      true,
				ConflictsWith: []string{"system"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sub_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"followed": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"job_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	// Unit File State
	systemdEnabled        = "enabled"
	systemdEnabledRuntime = "enabled-runtime"
	systemdLinked         = "linked"
	systemdLinkedRuntime  = "linked-runtime"
	systemdMasked         = "masked"
	systemdMaskedRuntime  = "masked-runtime"
	systemdStatic         = "static"
	systemdDisabled       = "disabled"
	systemdInvalid        = "invalid"
	systemdAlias          = "alias"     // not in doc
	systemdIndirect       = "indirect"  // not in doc
	systemdGenerated      = "generated" // not in doc

	// Load State
	systemdLoaded = "loaded"
	systemdError  = "error"
	//systemdMasked   = "masked"
	systemdNotFound = "not-found"

	// Active State
	systemdActive       = "active"
	systemdReloading    = "reloading"
	systemdInactive     = "inactive"
	systemdFailed       = "failed"
	systemdActivating   = "activating"
	systemdDeactivating = "deactivating"

	// Sub State (not exhaustive)
	systemdDead    = "dead"
	systemdRunning = "running"
)

func sdIsActive(active_state string) bool {
	return active_state == systemdActive || active_state == systemdReloading
}

func sdIsFailed(active_state string) bool {
	return active_state == systemdFailed
}

func sdMaskString(enable bool) string {
	if enable {
		return "mask"
	} else {
		return "unmask"
	}
}

func sdEnableString(enable bool) string {
	if enable {
		return "enable"
	} else {
		return "disable"
	}
}

func sdStartString(start bool) string {
	if start {
		return "start"
	} else {
		return "stop"
	}
}

func sdIsEnabled(unit_file_state string) bool {
	switch unit_file_state {
	case systemdEnabled, systemdEnabledRuntime, systemdStatic, systemdAlias, systemdIndirect, systemdGenerated:
		return true
	default:
		return false
	}
}

func sdIsMasked(unit_file_state string) bool {
	switch unit_file_state {
	case systemdMasked, systemdMaskedRuntime:
		return true
	default:
		return false
	}
}

func sdUser(ctx context.Context, m *providerConfiguration) (*systemd.Conn, error) {
	return systemd.NewUserConnectionContext(ctx)
}

func sdSystem(ctx context.Context, m *providerConfiguration) (*systemd.Conn, error) {
	return systemd.NewSystemConnectionContext(ctx)
}

func sdConn(ctx context.Context, d *schema.ResourceData, m interface{}) (*systemd.Conn, error) {
	user := d.Get("user").(bool)
	if user {
		return sdUser(ctx, m.(*providerConfiguration))
	} else {
		return sdSystem(ctx, m.(*providerConfiguration))
	}
}

func sdUnitLock(m interface{}, unit string) sync.Locker {
	c := m.(*providerConfiguration)
	var lock sync.Locker

	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.SdLocks == nil {
		lock = &sync.Mutex{}
		c.SdLocks = map[string]sync.Locker{
			unit: lock,
		}
	} else if mutex, ok := c.SdLocks[unit]; ok {
		lock = mutex
	} else {
		lock = &sync.Mutex{}
		c.SdLocks[unit] = lock
	}

	return lock
}

func resourceSystemdUnitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	return resourceSystemdUnitReadUnlocked(ctx, d, m)
}

func resourceSystemdUnitReadUnlocked(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	var errs diag.Diagnostics
	unit := d.Get("name").(string)

	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	statuses, err := sd.ListUnitsByNamesContext(ctx, []string{unit})
	if err != nil || len(statuses) < 1 {
		return diag.Errorf("cannot query unit %s: %v", unit, err)
	}
	status := statuses[0]

	d.Set("description", status.Description)
	d.Set("load_state", status.LoadState)
	d.Set("active_state", status.ActiveState)
	d.Set("sub_state", status.SubState)
	d.Set("followed", status.Followed)
	d.Set("job_id", status.JobId)
	d.Set("job_type", status.JobType)

	var rollback = map[string]interface{}{
		"exists":       strconv.FormatBool(status.LoadState != systemdNotFound),
		"load_state":   status.LoadState,
		"active_state": status.ActiveState,
		"sub_state":    status.SubState,
	}

	if status.LoadState == systemdNotFound {
		d.SetId("")
	} else {
		d.SetId(status.Name)

		unitFileState, err := sd.GetUnitFileStateContext(ctx, status.Name)
		if err != nil {
			return diag.Errorf("cannot get unit file state for %s: %v", status.Name, err)
		}

		enabled := sdIsEnabled(unitFileState)
		active := sdIsActive(status.ActiveState)
		rollback["active"] = strconv.FormatBool(active)
		rollback["enabled"] = strconv.FormatBool(enabled)
		rollback["unit_file_state"] = unitFileState

		d.Set("start", active)
		d.Set("enable", enabled)
	}

	if _, ok := d.GetOk("rollback"); !ok {
		err := d.Set("rollback", rollback)
		if err != nil {
			errs = append(errs, diag.FromErr(err)...)
		}
	}

	return errs
}

func resourceSystemdUnitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	errs := resourceSystemdUnitReadUnlocked(ctx, d, m)
	if errs != nil {
		return errs
	}

	return resourceSystemdUnitUpdate(ctx, d, m)
}

func resourceSystemdUnitDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_active := parseBoolDef(rollback["active"], false)
	rollback_enable := parseBoolDef(rollback["enabled"], false)

	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	// If unknown by systemd, there is nothing to rollback
	if d.Id() == "" {
		return nil
	}

	err = resourceSystemdEnable(ctx, d, sd, rollback_enable)
	if err != nil {
		return diag.Errorf("cannot %s unit %s: %v", sdEnableString(rollback_enable), unit, err)
	}

	restart := d.HasChange("restart_on")

	err = resourceSystemdActivate(ctx, d, sd, rollback_active, restart)
	if err != nil {
		return diag.Errorf("cannot %s unit %s: %v", sdStartString(rollback_active), unit, err)
	}

	errs := resourceSystemdUnitReadUnlocked(ctx, d, m)
	if errs != nil {
		return errs
	}

	return nil
}

func resourceSystemdEnable(ctx context.Context, d *schema.ResourceData, sd *systemd.Conn, enable bool) error {
	unit := d.Get("name").(string)
	unitFileState, err := sd.GetUnitFileStateContext(ctx, unit)
	if err != nil {
		return fmt.Errorf("cannot get unit file state for %s: %v", unit, err)
	}

	is_enabled := sdIsEnabled(unitFileState)

	if !is_enabled && enable {
		_, _, err = sd.EnableUnitFilesContext(ctx, []string{unit}, false, true)
	} else if is_enabled && !enable {
		_, err = sd.DisableUnitFilesContext(ctx, []string{unit}, false)
	}

	return err
}

func resourceSystemdMask(ctx context.Context, d *schema.ResourceData, sd *systemd.Conn, maskState string) error {
	unit := d.Get("name").(string)
	unitFileState, err := sd.GetUnitFileStateContext(ctx, unit)
	if err != nil {
		return fmt.Errorf("cannot get unit file state for %s: %v", unit, err)
	}

	if maskState != unitFileState && sdIsMasked(maskState) {
		_, err = sd.MaskUnitFilesContext(ctx, []string{unit}, maskState == systemdMaskedRuntime, true)
	} else if sdIsMasked(unitFileState) && !sdIsMasked(maskState) {
		_, err = sd.UnmaskUnitFilesContext(ctx, []string{unit}, false)
	}

	return err
}

func resourceSystemdActivate(ctx context.Context, d *schema.ResourceData, sd *systemd.Conn, activate, restart bool) error {
	unit := d.Get("name").(string)
	statuses, err := sd.ListUnitsByNamesContext(ctx, []string{unit})
	if err != nil || len(statuses) < 1 {
		return fmt.Errorf("cannot query unit %s: %v", unit, err)
	}
	status := statuses[0]

	is_active := sdIsActive(status.LoadState)

	complete := make(chan string)

	if restart && activate {
		_, err = sd.RestartUnitContext(ctx, unit, "replace", complete)
	} else if !is_active && activate {
		_, err = sd.StartUnitContext(ctx, unit, "replace", complete)
	} else if is_active && !activate {
		_, err = sd.StopUnitContext(ctx, unit, "replace", complete)
	}
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case completeStatus := <-complete:
		if completeStatus != "done" {
			return fmt.Errorf("Failed to activate %s: %s", unit, completeStatus)
		}
	}

	return err
}

func resourceSystemdUnitUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	start, has_start := d.GetOkExists("start")
	enable, has_enable := d.GetOkExists("enable")
	mask, has_mask := d.GetOkExists("enable")

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_active := parseBoolDef(rollback["active"], false)
	rollback_enable := parseBoolDef(rollback["enabled"], false)
	rollback_load_state := rollback["load_state"]

	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	if enable != nil && has_enable && d.HasChange("enable") {
		err = resourceSystemdEnable(ctx, d, sd, enable.(bool))
		if err != nil {
			return diag.Errorf("cannot %s unit %s: %v", sdEnableString(enable.(bool)), unit, err)
		}
	} else if !has_enable || enable == nil {
		err = resourceSystemdEnable(ctx, d, sd, rollback_enable)
		if err != nil {
			return diag.Errorf("cannot rollback %s unit %s: %v", sdEnableString(rollback_enable), unit, err)
		}
	}

	if mask != nil && has_mask && d.HasChange("mask") {
		var maskState string
		if mask.(bool) {
			maskState = systemdMasked
		}
		err = resourceSystemdMask(ctx, d, sd, maskState)
		if err != nil {
			return diag.Errorf("cannot %s unit %s: %v", sdMaskString(mask.(bool)), unit, err)
		}
	} else if rollback_load_state != nil && (!has_mask || mask == nil) {
		err = resourceSystemdMask(ctx, d, sd, rollback_load_state.(string))
		if err != nil {
			return diag.Errorf("cannot rollback %s unit %s: %v", sdMaskString(sdIsMasked(rollback_load_state.(string))), unit, err)
		}
	}

	restart := d.HasChange("restart_on")

	if start != nil && has_start && (d.HasChange("start") || restart) {
		err = resourceSystemdActivate(ctx, d, sd, start.(bool), restart)
		if err != nil {
			return diag.Errorf("cannot start unit %s: %v", sdStartString(start.(bool)), unit, err)
		}
	} else if !has_start || start == nil {
		err = resourceSystemdActivate(ctx, d, sd, rollback_active, restart)
		if err != nil {
			return diag.Errorf("cannot rollback %s unit %s: %v", sdStartString(rollback_active), unit, err)
		}
	}

	errs := resourceSystemdUnitReadUnlocked(ctx, d, m)
	if errs != nil {
		return errs
	}

	return nil
}
