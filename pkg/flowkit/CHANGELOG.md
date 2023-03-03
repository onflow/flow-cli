# Flowkit Changelog

All notable changes to flowkit APIs will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

---
The `Network` property was removed from `Contract` type. The network is now included in 
the `Aliases` on the contract. We also removed having multiple contracts by same name just to 
accommodate multiple aliases. Now there's only one contract identified by name, 
and if there are multiple network aliases they are contained in the `Aliases` list. 
- Package: `config`
- Type: `Contracts`

---
A method `Contracts.AddOrUpdate(name, contract)` was changed to not include the name, as it's 
already part of the contract you are adding.
- Method: `AddOrUpdate`
- Package: `config`
- Type: `Contracts`

---
Don't return error if contract by name not found but rather just a `nil`.
- Method: `ByName`
- Package: `config`
- Type: `Contracts`

### Added

---

New type `Aliases` was added to `Contracts`. 
Aliases contain new functions to get the aliases by network and add new aliases.
- Package: `config`
- Type: `Contracts`


---

`WithLogger` now takes zerolog instead of Logrus since that is what flow-emulator has changed to.
- Package: `gateway`
- Type: `EmulatorGateway`
