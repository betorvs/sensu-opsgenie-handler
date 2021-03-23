# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.0.2] - 2021-03-23
### Changed
- flag `--team` is not required anymore.

### Added
- flags `--schedule-team` and `--escalation-team` to add Responders field more correctly. From Opsgenie documentation ` If the API Key belongs to a team integration, this field will be overwritten with the owner team. ` [opsgenie api docs](https://docs.opsgenie.com/docs/alert-api#create-alert)
- flag `--addHooksToDetails` to add event.check.hooks.[].labels, event.check.hooks.[].command and event.check.hooks.[].output using event.check.hooks.[].name as key.

## [1.0.1] - 2021-03-19
### Added
- flag `--titlePrettify` to apply strings.Title in message title and remove these characters -, /, \

### Changed
- update golang to 1.15

## [1.0.0] - 2020-11-20

### Removed
- Removed flag `OPSGENIE_APIURL` now we use constants from opsgenie sdk-v2.
- Removed `opsgenie_priority` annotation. Should use: `"sensu.io/plugins/sensu-opsgenie-handler/config/priority": "P3"`.
- Removed travis-ci integration

### Changed
- Added flag `--region` to choose opsgenie region. Can be configured using environment variable too `OPSGENIE_REGION`. This feature replaces old `OPSGENIE_APIURL`.
- Added flag `--priority` to change Opsgenie default priority. String field. Expected: "P1", "P2", "P3", "P4" and "P5".
- Updated golang version to 1.14. As require for this updated golangci-lint to `v1.23.8`
- Changed opsgenie sdk to [v2](https://github.com/opsgenie/opsgenie-go-sdk-v2). 
- Changed from spf13/cobra to sensu-community/sensu-plugin-sdk. 
- Changed withAnnotations to parse all annotations, and exclude if it contains `sensu.io/plugins/sensu-opsgenie-handler/config`, to send to Opsgenie.

### Added
- Added more tests
- Added `--allowLabels` to parse all Labels and send to Opsgenie.
- Added `--fullDetails` to add all kind of details in Opsgenie.
- Added `--descriptionLimit`, `--descriptionTemplate`, `--includeEventInNote`, `--messageLimit`, `--messageTemplate` from this [forked](https://github.com/nixwiz/sensu-opsgenie-handler)

### Removed
- Removed flag `--allowOverride` to enable change opsgenie auth token and team. With this feature you can avoid creating multiples handlers only because one check or entity. We can achieve the same feature with Argument Annotations we can overwrite it even from a event through agent event api.

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