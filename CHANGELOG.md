# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

### Changed
- Move to Go Modules
- Use new Sensu SDK
- Move from Travis to GitHub Actions
- Reorganize README
- Added timestamps to all the test events

### Added
- Made message a configurable template, including a length limit

## [0.1.0] - 2020-03-31

### Changed
- Changed behaviour from opsgenie-handler to not add new alerts as a note and send them as alert.

### Removed
- Removed addNote func
- Removed eventKey func
- Removed eventTags func
- Removed goreleaser goos freeebsd and arch arm64

## [0.0.10] - 2020-02-23

### Added
- Parse Entity.Name, Check.Name, Entity.Namespace, Entity.Class as tags to Opsgenie
- Add OPSGENIE_SENSU_DASHBOARD variable to add new field in Description with "source: Sensu Dashboard URL"


## [0.0.9] - 2020-02-08

### Added
- Added the Sensu event json dump to the OpsGenie `Note` field.
- Added more tests

## [0.0.8] - 2020-01-20

### Changed
- change from dep to go mod
- gometalinter to golangci-lint
- correct goreleaser

## [0.0.7] - 2019-12-09

### Added
- Correct issue [#6](https://github.com/betorvs/sensu-opsgenie-handler/issues/6): `trim additional ending slash in --url argument`
- add script test.all.events.sh

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
