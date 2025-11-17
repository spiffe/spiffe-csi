# Changelog

## [0.2.9] - 2025-11-17

### Changed

- Dependency updates

## [0.2.8] - 2025-08-19

### Changed

- Dependency updates

## [0.2.7] - 2025-03-20

### Changed

- Dependency updates

## [0.2.6] - 2024-04-10

### Security

- Updated Golang to 1.22.2 and google.golang.org/grpc to v1.63.2 in order to address CVE-2023-45288 (#181)

## [0.2.5] - 2024-01-23

### Fixed

- The unmount operation now allows pods to terminate properly after a node reboot (#161)

## [0.2.4] - 2023-11-02

### Security

- Updated to google.golang.org/grpc v1.59.0 to address CVE-2023-44487

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
