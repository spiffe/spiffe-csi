# Changelog

## [0.2.0] - 2022-09-08

### Changed

- (**Breaking**) The driver now enforces that the CSI volume is declared read-only in the PodSpec
- The volume is now mounted R/W on the host to allow file attributes to be modified (e.g. SELinux)

## [0.1.0] - 2021-12-03

- First official release!
