## ⬆️ Install or Upgrade

Follow the [Flow CLI installation guide](https://docs.onflow.org/flow-cli/install/) for instructions on how to install or upgrade the CLI.

## 💥 Breaking Changes

### Configuration Format

The default configuration format (i.e. the contents of `flow.json`) has been updated.
It is now unified to work with all CLI commands. 
The new format is not backwards compatible with the old format.

If needed, you can generate a new configuration file with the `flow init` command.

Read more about the new configuration format in the [documentation](https://docs.onflow.org/flow-cli/configuration).

### Updated: `flow blocks get`

The `--latest`, `--id` and `--height` have been removed.

Instead, use the new argument syntax:

```sh
# get latest block
flow blocks get latest

# get a block by ID
flow blocks get 6bb0e0fceef9225a3cf9ceb6df9a31bd0063e6ee8e8dd7fdd93b831783243cd3

# get a block by height
flow blocks get 28329914
```

Read more about this change in the [documentation](https://docs.onflow.org/flow-cli/get-blocks).

### Removed: `flow keys decode`

The `flow keys decode` command has been temporarily removed due to a bug that requires further investigation.

### Removed: `flow keys save`

The `flow keys save` command has been removed in favour of an upcoming `flow accounts add` command. 

## ⚠️ Deprecation Warnings

The following functionality has been deprecated and will be removed in an upcoming release.

**`flow accounts create`, `flow accounts add-contract`, `flow accounts remove-contract`, `flow accounts update-contract`**

- Flag `--results` is deprecated, results are displayed by default.

**`flow accounts get`**

- Flag `--code` is deprecated, use `--contracts` flag instead.

**`flow events get`**

- Flag `--verbose` is deprecated.

**`flow keys generate`**

- Flag `--algo` is deprecated, use flag `--sig-algo`.

**`flow transactions send`**

- Flag `--code` is deprecated, use filename argument instead.
- Flag `--args` is deprecated, use `--arg` or `--args-json` instead.
- Flag `--results` is deprecated, results are displayed by default.

**`flow scripts execute`**

- Flag `--code` is deprecated, use filename argument instead.
- Flag `--args` is deprecated, use `--arg` or `--args-json` instead.

**`flow transactions status`**

- This command has been deprecated in favour of `flow transactions get`.

**`flow project init`**

- This command has been deprecated in favour of `flow init`.

**`flow project start-emulator`**

- This command has been deprecated in favour of `flow emulator`.

**`flow emulator start`**

- This command has been deprecated in favour of `flow emulator`.
