# Testing


## Mocks
Mocks are generated using [mockery](https://github.com/vektra/mockery).

To regenerate gateway mock go to the gateway directory `/pkg/flowkit/gateway` and run the command:
```shell
mockery --name=Gateway --output ../../../tests/mocks 
```
