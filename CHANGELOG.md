# Changelog

## [0.2.3] - 2023-02-24

### Changed

- Fixed a broken version string causing panic on startup.

## [0.2.2] - 2023-02-24

### Added

- Flag to configure the plugin name. This allows multiple instances of the driver to run under different names (#86)

### Changed

- Docker images are now multiarch with amd64 and arm64 support (#70)
- Docker images are now signed by sigstore (#73)

## [0.2.1] - 2022-11-07

### Changed

- Updated dependencies to quell false positive vulnerability reports (#58)

## [0.2.0] - 2022-09-08

### Changed

- (**Breaking**) The driver now enforces that the CSI volume is declared read-only in the PodSpec
- The volume is now mounted R/W on the host to allow file attributes to be modified (e.g. SELinux)

## [0.1.0] - 2021-12-03

- First official release!
