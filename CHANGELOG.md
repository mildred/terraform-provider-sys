## 1.3.29

* Fix broken tag 1.3.28

## 1.3.28

* rebuild documentation

## 1.3.27

* sys_shell_script: have a make script that avoids reimplementing read in the
  create script

## 1.3.26

* data source sys_os_releaase: parse /etc/os-release

## 1.3.25

* sys_file: fix directory permissions (you may need to recreate the resources)

## 1.3.24

* data source sys_shell_script: read data from a shell script
* data source sys_error: trigger an error

## 1.3.23

* sys_systemd_unit: fix unit active state not detected properly
* fix strings and add debug

## 1.3.22

* sys_systemd_unit: fix deadlock in unit activation (appears on rollback)

## 1.3.21

* sys_systemd_unit: fix creation

## 1.3.20

* sys_systemd_unit: fix masked

## 1.3.18

* sys_systemd_unit: does not produces a diff when start/enable/mask is unset

## 1.3.17

* sys_file: defaults to not executable

## 1.3.16

* sys_file: add ability to mask units

## 1.3.15

* lock systemd units by unit name to prevent conflicts in case unit is modified in multiple parts of the code

## 1.3.14

* Fix crash in sys_file
* Add sys_package.target_release

## 1.3.13

* sys_file: allow copying directories

## 1.3.8

* sys_file: implement symlink_destination

## 1.3.7

* sys_file: implement clear_destination

## 1.3.1

* sys_service: use dBus systemd API instead of shell commands
