package sys

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	systemd "github.com/coreos/go-systemd/v22/dbus"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// https://www.freedesktop.org/wiki/Software/systemd/dbus/

func resourceSystemdUnit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSystemdUnitCreate,
		ReadContext:   resourceSystemdUnitRead,
		DeleteContext: resourceSystemdUnitDelete,
		UpdateContext: resourceSystemdUnitUpdate,

		Description: "Handles a systemd unit with the dBus API.",

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "systemd unit name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enable": {
				Description:      "Enable the unit",
				Type:             schema.TypeBool,
				Optional:         true,
				DiffSuppressFunc: diffSuppressIfNil,
			},
			"mask": {
				Description:      "Mask the unit",
				Type:             schema.TypeBool,
				Optional:         true,
				DiffSuppressFunc: diffSuppressIfNil,
			},
			"start": {
				Description:      "Start the unit",
				Type:             schema.TypeBool,
				Optional:         true,
				DiffSuppressFunc: diffSuppressIfNil,
			},
			"restart_on": {
				Description: "Restart unit if this changes",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"rollback": {
				Description: "Rollback information to restore once the unit is destroyed",
				Type:        schema.TypeMap,
				Computed:    true,
			},
			"system": {
				Description:   "Uses the system systemd socket",
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"user"},
			},
			"user": {
				Description:   "Uses the user systemd socket",
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"system"},
			},
			"description": {
				Description: "Unit description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"load_state": {
				Description: "Unit load state",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"active_state": {
				Description: "Unit active state",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"sub_state": {
				Description: "Unit sub-state (specific to the unit type)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"followed": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_id": {
				Description: "Internal systemd job id",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"job_type": {
				Description: "Internal systemd job type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ignore_errors": {
				Description: "Ignore errors",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
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

func diffSuppressIfNil(k, old, new string, d *schema.ResourceData) bool {
	_, has_value := d.GetOkExists(k)
	return !has_value
}

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

func sdIsEnabled(unit_file_state string) (bool, bool) {
	switch unit_file_state {
	case systemdStatic, systemdEnabledRuntime, systemdAlias, systemdIndirect, systemdGenerated:
		return true, false
	case systemdEnabled:
		return true, true
	default:
		return false, true
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

func withSeverity(d *schema.ResourceData, errs diag.Diagnostics) diag.Diagnostics {
	if d.Get("ignore_errors").(bool) {
		for i := range errs {
			errs[i].Severity = diag.Warning
		}
	}
	return errs
}

func resourceSystemdUnitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	log.Printf("[DEBUG] About to read %s\n", unit)
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

		enabled, enableable := sdIsEnabled(unitFileState)
		active := sdIsActive(status.ActiveState)
		masked := sdIsMasked(status.LoadState)
		rollback["active"] = strconv.FormatBool(active)
		rollback["enabled"] = strconv.FormatBool(enabled)
		rollback["masked"] = strconv.FormatBool(masked)
		rollback["unit_file_state"] = unitFileState

		if _, has_start := d.GetOkExists("start"); has_start {
			d.Set("start", active)
		}
		if _, has_enable := d.GetOkExists("enable"); has_enable && enableable {
			d.Set("enable", enabled)
		}
		if _, has_mask := d.GetOkExists("mask"); has_mask {
			d.Set("mask", masked)
		}
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
	log.Printf("[DEBUG] About to create %s\n", unit)
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

	return resourceSystemdUnitUpdateUnlocked(ctx, d, m, true)
}

func resourceSystemdUnitDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	unit := d.Get("name").(string)
	log.Printf("[DEBUG] About to delete %s\n", unit)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	log.Printf("[DEBUG] connect to systemd\n")
	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_active := parseBoolDef(rollback["active"], false)
	rollback_enable := parseBoolDef(rollback["enabled"], false)

	log.Printf("[DEBUG] systemctl daemon-reload\n")
	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	// If unknown by systemd, there is nothing to rollback
	if d.Id() == "" {
		log.Printf("[DEBUG] Deleted %s (no rollback)\n", unit)
		return nil
	}

	log.Printf("[DEBUG] Rollback %s %s\n", sdEnableString(rollback_enable), unit)
	err = resourceSystemdEnable(ctx, d, sd, rollback_enable)
	if err != nil {
		derr := diag.Errorf("cannot %s unit %s: %v", sdEnableString(rollback_enable), unit, err)
		return withSeverity(d, derr)
	}

	restart := d.HasChange("restart_on")

	log.Printf("[DEBUG] Rollback %s %s (restart: %v)\n", sdStartString(rollback_active), unit, restart)
	err = resourceSystemdActivate(ctx, d, sd, rollback_active, restart)
	if err != nil {
		derr := diag.Errorf("cannot %s unit %s: %v", sdStartString(rollback_active), unit, err)
		return withSeverity(d, derr)
	}

	log.Printf("[DEBUG] Read unit %s after rollback\n", unit)
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

	is_enabled, is_enableable := sdIsEnabled(unitFileState)

	if is_enableable && !is_enabled && enable {
		log.Printf("[TRACE] Enable %s (enable=%v, is_enabled=%v, is_enableable=%v)\n", unit, enable, is_enabled, is_enableable)
		_, _, err = sd.EnableUnitFilesContext(ctx, []string{unit}, false, true)
	} else if is_enableable && is_enabled && !enable {
		log.Printf("[TRACE] Disasble %s (enable=%v, is_enabled=%v, is_enableable=%v)\n", unit, enable, is_enabled, is_enableable)
		_, err = sd.DisableUnitFilesContext(ctx, []string{unit}, false)
	} else {
		log.Printf("[TRACE] Do not enable %s (enable=%v, is_enabled=%v, is_enableable=%v)\n", unit, enable, is_enabled, is_enableable)
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
		log.Printf("[TRACE] Mask (%s) %s (state=%v, is_masked=%v, do_mask=%v)\n", maskState, unit, unitFileState, sdIsMasked(maskState), sdIsMasked(maskState))
		_, err = sd.MaskUnitFilesContext(ctx, []string{unit}, maskState == systemdMaskedRuntime, true)
	} else if sdIsMasked(unitFileState) && !sdIsMasked(maskState) {
		log.Printf("[TRACE] Unmask (%s) %s (state=%v, is_masked=%v, do_mask=%v)\n", maskState, unit, unitFileState, sdIsMasked(maskState), sdIsMasked(maskState))
		_, err = sd.UnmaskUnitFilesContext(ctx, []string{unit}, false)
	} else {
		log.Printf("[TRACE] Do not mask (%s) %s (state=%v, is_masked=%v, do_mask=%v)\n", maskState, unit, unitFileState, sdIsMasked(maskState), sdIsMasked(maskState))
	}

	return err
}

func resourceSystemdActivate(ctx context.Context, d *schema.ResourceData, sd *systemd.Conn, activate, restart bool) error {
	unit := d.Get("name").(string)
	statuses, err := sd.ListUnitsByNamesContext(ctx, []string{unit})
	if err != nil || len(statuses) < 1 {
		return fmt.Errorf("cannot query unit %s: %v", unit, err)
	}

	log.Printf("[TRACE] Activate %v %s (restart: %v): statuses = %v\n", activate, unit, restart, statuses)

	status := statuses[0]
	is_active := sdIsActive(status.ActiveState)

	log.Printf("[TRACE] Activate %v %s (restart: %v) active=%s is_active=%v activate=%v\n", activate, unit, restart, status.ActiveState, is_active, activate)

	complete := make(chan string)

	if restart && activate {
		log.Printf("[TRACE] Activate %v %s (restart: %v): systemctl restart %s\n", activate, unit, restart, unit)
		_, err = sd.RestartUnitContext(ctx, unit, "replace", complete)
	} else if !is_active && activate {
		log.Printf("[TRACE] Activate %v %s (restart: %v): systemctl start %s\n", activate, unit, restart, unit)
		_, err = sd.StartUnitContext(ctx, unit, "replace", complete)
	} else if is_active && !activate {
		log.Printf("[TRACE] Activate %v %s (restart: %v): systemctl stop %s\n", activate, unit, restart, unit)
		_, err = sd.StopUnitContext(ctx, unit, "replace", complete)
	} else {
		log.Printf("[TRACE] Activate %v %s (restart: %v): nothing to do\n", activate, unit, restart)
		close(complete)
		return nil
	}
	if err != nil {
		return err
	}

	log.Printf("[TRACE] Activate %v %s (restart: %v): wait for complete\n", activate, unit, restart)
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
	log.Printf("[DEBUG] About to update %s\n", unit)
	lock := sdUnitLock(m, unit)
	lock.Lock()
	defer lock.Unlock()

	return resourceSystemdUnitUpdateUnlocked(ctx, d, m, false)
}

func resourceSystemdUnitUpdateUnlocked(ctx context.Context, d *schema.ResourceData, m interface{}, creating bool) diag.Diagnostics {
	unit := d.Get("name").(string)

	sd, err := sdConn(ctx, d, m)
	if err != nil {
		return diag.Errorf("cannot connect to systemd: %v", err)
	}

	defer sd.Close()

	start, has_start := d.GetOkExists("start")
	enable, has_enable := d.GetOkExists("enable")
	mask, has_mask := d.GetOkExists("mask")

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_active     := parseBoolDef(rollback["active"], false)
	rollback_enable     := parseBoolDef(rollback["enabled"], false)
	rollback_load_state := rollback["load_state"]
	rollback_file_state := rollback["unit_file_state"]

	err = sd.ReloadContext(ctx)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	log.Printf("[TRACE] Update %s start=%v has_start=%v enable=%v has_enable=%v mask=%v has_mask=%v rollback_active=%v rollback_enable=%v, rollback_load_state=%v, rollback_file_state=%v\n",
		unit, start, has_start, enable, has_enable, mask, has_mask, rollback_active, rollback_enable, rollback_load_state, rollback_file_state)

	if enable != nil && has_enable && (creating || d.HasChange("enable")) {
		err = resourceSystemdEnable(ctx, d, sd, enable.(bool))
		if err != nil {
			derr := diag.Errorf("cannot %s unit %s: %v", sdEnableString(enable.(bool)), unit, err)
			return withSeverity(d, derr)
		}
	} else if !has_enable || enable == nil {
		err = resourceSystemdEnable(ctx, d, sd, rollback_enable)
		if err != nil {
			derr := diag.Errorf("cannot rollback %s unit %s: %v", sdEnableString(rollback_enable), unit, err)
			return withSeverity(d, derr)
		}
	}

	if mask != nil && has_mask && (creating || d.HasChange("mask")) {
		var maskState string
		if mask.(bool) {
			maskState = systemdMasked
		}
		err = resourceSystemdMask(ctx, d, sd, maskState)
		if err != nil {
			derr := diag.Errorf("cannot %s unit %s: %v", sdMaskString(mask.(bool)), unit, err)
			return withSeverity(d, derr)
		}
	} else if rollback_load_state != nil && (!has_mask || mask == nil) {
		err = resourceSystemdMask(ctx, d, sd, rollback_load_state.(string))
		if err != nil {
			derr := diag.Errorf("cannot rollback %s unit %s: %v", sdMaskString(sdIsMasked(rollback_load_state.(string))), unit, err)
			return withSeverity(d, derr)
		}
	}

	restart := d.HasChange("restart_on")
	log.Printf("[TRACE] Update %s restart=%v\n", unit, restart)

	if start != nil && has_start && (creating || d.HasChange("start") || restart) {
		err = resourceSystemdActivate(ctx, d, sd, start.(bool), restart)
		if err != nil {
			derr := diag.Errorf("cannot %s unit %s: %v", sdStartString(start.(bool)), unit, err)
			return withSeverity(d, derr)
		}
	} else if !has_start || start == nil {
		err = resourceSystemdActivate(ctx, d, sd, rollback_active, restart)
		if err != nil {
			derr := diag.Errorf("cannot rollback %s unit %s: %v", sdStartString(rollback_active), unit, err)
			return withSeverity(d, derr)
		}
	}

	errs := resourceSystemdUnitReadUnlocked(ctx, d, m)
	if errs != nil {
		return errs
	}

	return nil
}
