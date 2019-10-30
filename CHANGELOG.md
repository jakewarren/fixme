# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2019-10-30
### Added
- [#3](https://github.com/jakewarren/fixme/pull/3) Added support for LaTeX comments. Thanks [@xarantolus](https://github.com/xarantolus)! 

### Changed
- Migrated to go modules.
- Build releases with goreleaser.

## [1.2.0] - 2019-07-22
### Fixed
- [#2](https://github.com/jakewarren/fixme/pull/2) Fixed an issue with output being inconsistent between runs. Thanks [@reidab](https://github.com/reidab)!

### Changed
- If the user doesn't specify a directory, use the current directory as a default.
- Exclude the vendor directories by default.

## [1.1.0] - 2017-10-03
### Changed
- Added a max limit to the number of spawned goroutines to prevent trouble when running on large directories.

## [1.0.0] - 2017-10-03
- Initial release
