## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](https://docs.onflow.org/flow-cli/install/) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Initializing Configuration
Configuration format was unified to work for all CLI commands.
Generating new configuration is done by executing command `flow project init`.
Command `flow init` was removed as it provides the same functionality as project init command.

###⚠️ Deprecated Flags

**Accounts**
- Flag `--results` is deprecated, results are displayed by default.
- Flag `--code` on `accounts get` command was deprecated, use `--contracts` flag instead.

**Blocks**
- Flags `--latest`, `--id` and `--height` were deprecated in favour of using block argument.
  Command should be used with query argument where you can specify block height, id or value `latest`.
  Read more about it in the [documentation](https://docs.onflow.org/flow-cli/get-blocks).

**Events**
- Flag `--verbose` was deprecated.

**Keys**
- Flag `--algo` was renamed to `--sig-algo`.

**Transactions**
- Flag `--code` was deprecated, use filename argument instead.
- Flag `--results` was deprecated, results are displayed by default.

