## ⬆️ Install or Upgrade
Follow the Flow CLI installation guide for instructions on how to install or upgrade the CLI.

## 🐞 Bug Fixes

### Block Transaction IDs
Fetching a block didn't include transaction IDs when using the `--include transactions` flag due to a regression in the command layer.

### Script Execution Error
**Implemented by the community: @bjartek** 🙌

Script execution error wasn't returned when using hosted gateway
implementation.

### Parsing Boolean in Configuration
**Implemented by the community: @bluesign** 🙌

Parsing booleans in the flow configuration deployment arguments weren't working properly. This bugfix addresses that problem and allows you to pass boolean type in the `args` section such as:
```js
...
"args": [
    {"type": "Bool", "value": true}
]
...
```

### Cross-referencing Composed Configuration
Cross-referencing values in the composed configuration wasn't working correctly as the validation was done per configuration instead on the higher level on the composed configuration.

## 🛠 Improvements

### Arguments Without Types
**Implemented by the community: @bluesign** 🙌

Great improvement to the argument parsing. CLI now infers types from transaction parameters and script parameters, so it's not needed anymore to specify the type explicitly. This new improvement also supports passing arrays and dictionaries.

The new command format is:
```bash
flow scripts execute <filename> [<argument> <argument> ...] [flags]
```
Example:
```bash
> flow scripts execute script.cdc "Meow" "Woof"
```
In the example above the string, type is inferred from the script source code.

More complex example:
```bash
> flow transactions send tx1.cdc Foo 1 2 10.9 0x1 '[123,222]' '["a","b"]'
```
Transaction code:
```
transaction(a: String, b: Int, c: UInt16, d: UFix64, e: Address, f: [Int], g: [String]) {
	prepare(authorizer: AuthAccount) {}
}
```

### Idiomatic Accessors
**Implemented by the community: @bjartek** 🙌

Getter methods were rewritten in idiomatic Go, containing a value and an error. This solves some edge case bugs where the value is missing but the returned value is not
checked for nil.

### Configuration Loader Improvement
**Implemented by the community: @bjartek** 🙌

Configuration loading logic was improved, and it now only loads global configuration if local isn't present. Furthermore, it improves the `configuration add` commands that now only allow passing a single config by using the `-f` flag and it requires at least one local configuration to be present.
