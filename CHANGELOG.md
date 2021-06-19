## 1.3.18

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
