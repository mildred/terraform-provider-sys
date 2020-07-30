package sys

import (
	"context"
	"strconv"

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
				Type:          schema.TypeBool,
				Description:   "Enable the unit",
				Optional:      true,
				ConflictsWith: []string{"disable"},
			},
			"disable": {
				Type:          schema.TypeBool,
				Description:   "Disable the unit",
				Optional:      true,
				ConflictsWith: []string{"enable"},
			},
			"start": {
				Type:          schema.TypeBool,
				Description:   "Start the unit",
				Optional:      true,
				ConflictsWith: []string{"stop"},
			},
			"stop": {
				Type:          schema.TypeBool,
				Description:   "Stop the unit",
				Optional:      true,
				ConflictsWith: []string{"start"},
			},
			"restart_on": {
				Type:        schema.TypeMap,
				Description: "Restart unit if this changes",
				Optional:    true,
				ForceNew:    true,
			},
			"rollback": {
				Type:        schema.TypeMap,
				Description: "Rollback information to restore once the unit is destroyed",
				Computed:    true,
			},
		},
	}
}

func resourceSystemdUnitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log := m.(*providerConfiguration).Logger
	var errs diag.Diagnostics
	unit := d.Get("name").(string)

	err := systemdDaemonReload(log)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	exists, err := systemdIsExists(unit)
	if err != nil {
		errs = append(errs, diag.Errorf("error checking if unit exists: %v", err)...)
	} else if exists {
		d.SetId(unit)
	} else {
		d.SetId("")
	}

	var rollback = map[string]interface{}{
		"exists": strconv.FormatBool(exists),
	}

	if exists {

		active, err := systemdIsActive(unit)
		if err != nil {
			errs = append(errs, diag.Errorf("error running systemd is-active: %v", err)...)
		} else {
			if _, ok := d.GetOk("start"); ok {
				d.Set("start", active)
			}
			if _, ok := d.GetOk("stop"); ok {
				d.Set("stop", !active)
			}
			rollback["active"] = strconv.FormatBool(active)
		}

		enabled, err := systemdIsEnabled(unit)
		if err != nil {
			errs = append(errs, diag.Errorf("error running systemd is-enabled: %v", err)...)
		} else {
			if _, ok := d.GetOk("enable"); ok {
				d.Set("enable", enabled)
			}
			if _, ok := d.GetOk("disable"); ok {
				d.Set("disable", !enabled)
			}
			rollback["enabled"] = strconv.FormatBool(enabled)
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
	log := m.(*providerConfiguration).Logger
	err := systemdDaemonReload(log)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	errs := resourceSystemdUnitRead(ctx, d, m)
	if errs != nil {
		return errs
	}

	return resourceSystemdUnitUpdate(ctx, d, m)
}

func resourceSystemdUnitDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log := m.(*providerConfiguration).Logger
	unit := d.Get("name").(string)

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_start := parseBoolDef(rollback["active"], false)
	rollback_enable := parseBoolDef(rollback["enabled"], false)

	err := systemdDaemonReload(log)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	err = systemdStartStopEnableDisable(log, unit, rollback_start, !rollback_start, rollback_enable, !rollback_enable)
	if err != nil {
		return diag.Errorf("cannot delete: %v", err)
	}

	errs := resourceSystemdUnitRead(ctx, d, m)
	if errs != nil {
		return errs
	}

	return nil
}

func resourceSystemdUnitUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log := m.(*providerConfiguration).Logger
	unit := d.Get("name").(string)
	start := d.Get("start").(bool)
	stop := d.Get("stop").(bool)
	enable := d.Get("enable").(bool)
	disable := d.Get("disable").(bool)

	rollback := d.Get("rollback").(map[string]interface{})
	rollback_start := parseBoolDef(rollback["active"], false)
	rollback_enable := parseBoolDef(rollback["enabled"], false)

	if !start && !stop {
		start = rollback_start
		stop = !rollback_start
	}

	if !enable && !disable {
		enable = rollback_enable
		disable = !rollback_enable
	}

	err := systemdDaemonReload(log)
	if err != nil {
		return diag.Errorf("cannot reload systemd: %v", err)
	}

	err = systemdStartStopEnableDisable(log, unit, start, stop, enable, disable)
	if err != nil {
		return diag.Errorf("cannot start/enable unit: %v", err)
	}

	errs := resourceSystemdUnitRead(ctx, d, m)
	if errs != nil {
		return errs
	}

	return nil
}
