# ⬆️ Install or Upgrade
Follow the Flow CLI installation guide for instructions on how to install or upgrade the CLI.

# 🐞 Bug Fixes

## Non-existing Service Account
**Implemented by the community: @bjartek**
The bug crashing the CLI when referencing a non-existing service account was fixed.

## 🛠 Improvements

## Improved Event Command
**Implemented by the community: @bjartek**
Big improvements for the event commands. You can now specify multiple event names when you are fetching the events. The command will combine all the events together in the result.
```
flow events get A.1654653399040a61.FlowToken.TokensDeposited A.1654653399040a61.FlowToken.TokensWithdrawn
```

Command format was changed so it now requires flags for start block height (`--start`), end block height (`--end`), and allows a new flag for specifying the number of blocks since the last block height (`--last`). Some examples of usage:

```
flow events get A.1654653399040a61.FlowToken.TokensDeposited --start 11559500 --end 11559600
flow events get A.1654653399040a61.FlowToken.TokensDeposited --last 20 
```

The event fetching will be done concurrently using workers, the default worker count is 10, but you can specify the count explicitly using the `--workers` flag and also the number of blocks each worker fetches with `--batch` flag. This functionality brings great speed improvements and also allows you to fetch more blocks than the current limit.

## Add Mainnet Alias
**Implemented by the community: @bjartek**
The improved config command allows you to add new mainnet aliases. Example usage:
```
flow config add contract --mainnet-alias Alice 
```

## Continuous Delivery
Continuous delivery with help of Github actions.



