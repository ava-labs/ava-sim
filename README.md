<div align="center">
  <img src="resources/AvalancheLogoRed.png?raw=true">
</div>

# ava-sim
`ava-sim` makes it easy for anyone to spin up a local Avalanche network to use
standard APIs or to test a custom VM.

## Standard Network
`./scripts/run.sh`

```txt
standard VM endpoints now accessible at:
NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg: http://127.0.0.1:9650
NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ: http://127.0.0.1:9652
NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN: http://127.0.0.1:9654
NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu: http://127.0.0.1:9656
NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5: http://127.0.0.1:9658
```

## Custom VM (Subnet)
https://docs.avax.network/build/tutorials/platform/create-custom-blockchain
`./scripts/run.sh [vm] [vm-genesis]`

### Example: [TimestampVM](https://github.com/ava-labs/timestampvm)
`./scripts/subnet-example.sh`

```txt
NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg validating blockchain 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ validating blockchain 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN validating blockchain 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu validating blockchain 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5 validating blockchain 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
```

```txt
NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg bootstrapped 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ bootstrapped 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN bootstrapped 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu bootstrapped 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5 bootstrapped 28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
```

```txt
Custom VM endpoints now accessible at:
NodeID-7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg: http://127.0.0.1:9650/ext/bc/28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ: http://127.0.0.1:9652/ext/bc/28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN: http://127.0.0.1:9654/ext/bc/28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-GWPcbFJZFfZreETSoWjPimr846mXEKCtu: http://127.0.0.1:9656/ext/bc/28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5: http://127.0.0.1:9658/ext/bc/28TtJ7sdYvdgfj1CcXo5o3yXFMhKLrv4FQC9WhgSHgY6YNYRs2
```

## What this is NOT
This tool is not intended to be a full-fledged node configurator. Rather it is
more for testing interactions with a standard configuration and testing custom
VMs.

[avash](https://github.com/ava-labs/avash) provides similar functionality
to `ava-sim` but requires ...
