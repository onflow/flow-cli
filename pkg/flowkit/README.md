## Flow CLI

Flow CLI package contains all the functionality used in CLI. 
This package is meant to be used by third party (langauge server etc...).

### Config

Config package implements parsing and serializing configuration
and persisting state needed for some commands.

### Gateway

Gateway package contains functions that interact with the Flow blockchain. 
This package offers abstraction over communicating with the Flow network and 
takes care of initializing the network client and handling errors. 

Function accept arguments in go-sdk types or lib types and must already be validated.
Client is already initialized and only referenced inside here.

### Services

Service layer is meant to be used as an api. Service function accepts raw
arguments, validate them, use gateways to do network interactions and lib to
build resources needed in gateways.
