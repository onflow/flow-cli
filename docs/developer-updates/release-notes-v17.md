## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](https://docs.onflow.org/flow-cli/install/) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Initializing Configuration
Configuration format was unified to work for all CLI commands.
Generating new configuration is done by executing command `flow init`.

###⚠️  No longer supported flags

**Accounts**
- Flag `--results` is deprecated, results are displayed by default.
- Flag `--code` on `accounts get` command is deprecated, use `--contracts` flag instead.

**Blocks**
- Flags `--latest`, `--id` and `--height` are deprecated in favour of using block argument.
  Command should be used with query argument where you can specify block height, id or value `latest`.
  Read more about it in the [documentation](https://docs.onflow.org/flow-cli/get-blocks).

**Events**
- Flag `--verbose` is deprecated.

**Keys**
- Flag `--algo` is deprecated, use flag `--sig-algo`.

**Transactions**
- Flag `--code` is deprecated, use filename argument instead.
- Flag `--results` is deprecated, results are displayed by default.

