# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## [0.0.6] - 2019-11-24

### Added
- Add `OPSGENIE_ANNOTATIONS` to parse annotations to include that information inside the alert.
- Update README.

## [0.0.5] - 2019-10-15

### Added
- Add `OPSGENIE_APIURL` to change OpsGenie API URL
- Updated Gopkg.lock file.
- Changed travis go version.

## [0.0.4] - 2019-08-26

### Added
- Add bonsai configuration

## [0.0.3] - 2019-08-02

### Added
- Add OpsGenie Priority as annotations inside check annotations.
- Add Get, Close and Add Note functions to manage alerts already open. 

## [0.0.2] - 2019-07-10

### Added
- Add OpsGenie Priority as annotations inside sensu-agent to override default Alert Event Priority in OpsGenie.

## [0.0.1] - 2019-07-10

### Added
- Initial release