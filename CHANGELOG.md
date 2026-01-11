# Changelog

All notable changes to this project will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.0] - 2026-01-11

### Fixed

- Fixed reconnect handling for Home Assistant MQTT integration ([#9](https://github.com/joshuasing/pws_exporter/pull/9))

### Changed

- Updated to Go 1.25.5 ([#11](https://github.com/joshuasing/pws_exporter/pull/11), [#20](https://github.com/joshuasing/pws_exporter/pull/20))
- Updated dependencies ([#11](https://github.com/joshuasing/pws_exporter/pull/11))
- Switched to using `chainguard/static` base images ([#26](https://github.com/joshuasing/pws_exporter/pull/26))

### Removed

- Removed support for `linux/arm/v7` in Docker images ([#26](https://github.com/joshuasing/pws_exporter/pull/26))

-----

_Looking for the changelog for an older version? Older releases can be found at:
https://github.com/joshuasing/pws_exporter/releases_

[Unreleased]: https://github.com/joshuasing/pws_exporter/compare/v0.4.0...HEAD
[v0.4.0]: https://github.com/joshuasing/pws_exporter/releases/tag/v0.4.0
