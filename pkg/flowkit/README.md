## Flowkit Package Design
Flowkit is a core package used by the CLI commands. It features APIs for interacting with the Flow network 
in the context of flow.json configuration values. Flowkit is defined by the [interface here](./services.go).

Flowkit contains multiple subpackages, the most important ones are: 
- **config**: parsing and storing of flow.json values, as well as validation, 
- **gateway**: implementation of Flow AN methods, uses emulator as well as Go SDK to communicate with ANs,
- **project**: stateful operations on top of flow.json, which allows resolving imports in contracts used in deployments

It is important we define clear boundaries between flowkit and other CLI packages. If we are in doubt where certain 
methods should be implemented we must ask ourselves if the method provides value for any other consumers of the 
pacakge or only provides utility for CLI usage, if it's only providing utility for CLI then it should be added inside 
the internal package, instead of flowkit. If in doubt better to add things to internal package and then move to flowkit 
if such need is identified. 


