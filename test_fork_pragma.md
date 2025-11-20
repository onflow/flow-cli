# Fork Testing Pragma

## Overview

The `#test_fork` pragma is the **recommended way** to configure fork tests in Cadence. It configures tests to run against a snapshot of a live Flow network (mainnet or testnet), allowing you to test against real deployed contracts and production data without deploying anything to the live network.

**Why use the pragma?**

- ✅ **No CLI flags required** - Test configuration lives in your test file
- ✅ **Works with `flow test`** - No special commands needed
- ✅ **Self-documenting** - Other developers can see the test is forked just by reading the file
- ✅ **Mix test types freely** - Run local tests and fork tests with a single `flow test` command
- ✅ **No flow.json changes** - Test configuration stays with the test code

## Syntax

```cadence
#test_fork(network: "mainnet", height: nil)
```

Place the pragma at the top of your test file, before any imports.

## Parameters

### `network` (required)

The Flow network to fork from.

**Valid values:**

- `"mainnet"` - Fork from Flow mainnet
- `"testnet"` - Fork from Flow testnet
- Any network name defined in your `flow.json` networks configuration

### `height` (optional)

The block height to fork from.

**Valid values:**

- `nil` - Use the latest sealed block (recommended for most cases)
- `UInt64` - Fork from a specific historical block height (e.g., `height: 85229104`)

## Purpose

Fork testing enables:

1. **Integration testing** - Test interactions with real deployed contracts without deploying your own
2. **Account impersonation** - Execute transactions as any mainnet account without private keys
3. **Production state validation** - Verify behavior against actual on-chain data
4. **Historical debugging** - Test against blockchain state at a specific block height

## Examples

### Basic Fork Test (Latest Block)

```cadence
#test_fork(network: "mainnet", height: nil)

import Test
import "SomeMainnetContract"

access(all) fun testAgainstMainnet() {
    // Test code runs against mainnet state
    let account = Test.getAccount(0x1234567890abcdef)
    // All changes happen locally in test environment
}
```

### Fork Test at Specific Height

```cadence
#test_fork(network: "mainnet", height: 85229104)

import Test

access(all) fun testHistoricalState() {
    // Test against blockchain state at block 85229104
    // Useful for reproducing historical bugs or validating fixes
}
```

### Using Custom Network from flow.json

```cadence
#test_fork(network: "previewnet", height: nil)

import Test

access(all) fun testOnPreviewnet() {
    // Forks from custom network defined in your flow.json
}
```

## Running Fork Tests

The beauty of the `#test_fork` pragma is that you don't need any special flags:

```bash
# Run all tests - local AND fork tests work together!
flow test

# Run a specific fork test
flow test path/to/fork_test.cdc

# Run only local tests (exclude fork tests)
flow test --skip-fork
```

**No need for:**

- ❌ `flow test --fork mainnet` flags
- ❌ `flow test --fork-host` configuration
- ❌ Different commands for different test types

Just use `flow test` and the pragma handles everything!

## Best Practices

### Use the Pragma (Not CLI Flags)

**Recommended:**

```cadence
#test_fork(network: "mainnet", height: nil)
import Test
// Test code...
```

**Not Recommended:**

```bash
flow test --fork mainnet path/to/test.cdc  # Requires flags every time
```

The pragma keeps fork configuration with your test code where it belongs.

### Organize Your Tests

```
tests/
  ├── unit_tests/          # Local tests
  │   └── contract_test.cdc
  └── integration_tests/   # Fork tests
      └── mainnet_integration_test.cdc
```

Run everything with: `flow test`

## Notes

- Fork tests require network connectivity to access the forked network
- All mutations occur only in your local test environment
- Fork tests are slightly slower than local tests due to network state fetching
- The pragma makes it easy to mix fork tests and regular tests in the same test suite
- CI/CD friendly - no special configuration needed, just `flow test`
