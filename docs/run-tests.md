---
title: Run Cadence tests with the Flow CLI
sidebar_title: Run Cadence tests
description: How to run Cadence tests from the command line
---

The Flow CLI provides a command to run Cadence tests. 

```shell
flow test /path/to/test_script.cdc
```

⚠️ The `test` command expects configuration to be initialized. See [flow init](initialize-configuration.md) command.


## Example Usage
A simple Cadence script `test_script.cdc`, which has a test case for running a cadence script on-chain:
```cadence
import Test

pub fun testSimpleScript() {
    var blockchain = Test.newEmulatorBlockchain()
    var result = blockchain.executeScript(
        "pub fun main(a: Int, b: Int): Int { return a + b }",
        [2, 3]
    )
    
    assert(result.status == Test.ResultStatus.succeeded)
    assert((result.returnValue! as! Int) == 5)
}
```
Above test-script can be run with the CLI as follows, and the test results will be printed on the console.
```shell
> flow test test_script.cdc

Running tests...

Test results:
- PASS: testSimpleScript
```

To learn more about writing tests in Cadence, have a look at the [Cadence testing framework](https://developers.flow.com/cadence/testing-framework).

