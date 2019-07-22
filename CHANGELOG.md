# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2019-07-22
### Fixed
- Fixed an issue with output being inconsistent between runs. Thanks @reidab!

### Changed
- If the user doesn't specify a directory, use the current directory as a default.
- Exclude the vendor directories by default.

## [1.1.0] - 2017-10-03
### Changed
- Added a max limit to the number of spawned goroutines to prevent trouble when running on large directories.

## [1.0.0] - 2017-10-03
- Initial release
