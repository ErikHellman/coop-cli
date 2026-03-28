# coop-cli

[![Test](https://github.com/ErikHellman/coop-cli/actions/workflows/test.yml/badge.svg)](https://github.com/ErikHellman/coop-cli/actions/workflows/test.yml)
[![Release](https://github.com/ErikHellman/coop-cli/actions/workflows/release.yml/badge.svg)](https://github.com/ErikHellman/coop-cli/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ErikHellman/coop-cli)](https://goreportcard.com/report/github.com/ErikHellman/coop-cli)

A command-line tool for interacting with [Coop.se](https://www.coop.se) online grocery shopping. Search for products, manage your shopping cart, and find stores — all from the terminal.

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/ErikHellman/coop-cli/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ErikHellman/coop-cli/main/install.ps1 | iex
```

### From source

```bash
go install github.com/ErikHellman/coop-cli@latest
```

## Configuration

### Credentials

All commands except `stores` and `update` require a Coop account. Provide credentials via flags or environment variables:

```bash
# Flags
coop-cli --email user@example.com --password secret --store 251300 search "mjölk"

# Environment variables (recommended)
export COOP_EMAIL=user@example.com
export COOP_PASSWORD=secret
export COOP_STORE=251300
```

### Store ID

A store ID is required for product search and cart commands. Find your store first:

```bash
coop-cli stores Uppsala
```

Then use the `STORE ID` from the output with `--store` or `COOP_STORE`:

```bash
coop-cli --store 251300 search "kaffe"

# Or set it once via environment variable
export COOP_STORE=251300
```

## Usage

The examples below assume `COOP_EMAIL`, `COOP_PASSWORD`, and `COOP_STORE` are set as environment variables.

### Search for products

```bash
coop-cli search "mjölk"
coop-cli search "bröd" --limit 20
```

Output includes product ID, name, manufacturer, package size, price, comparison price, and category.

### Manage the shopping cart

```bash
coop-cli cart list                    # List cart contents with totals
coop-cli cart add 7300156573186       # Add product (quantity 1)
coop-cli cart add 7300156573186 3     # Add product with specific quantity
coop-cli cart remove 7300156573186    # Remove product from cart
coop-cli cart clear                   # Empty the entire cart
```

Adding and removing products prints the full cart after the operation.

### Find stores

```bash
coop-cli stores Uppsala
coop-cli stores Göteborg
```

Search by city, store name, or address. This command does not require login or a store ID.

### Update

```bash
coop-cli update       # Update to the latest release
coop-cli --version    # Show current version
```

## Releasing

Releases are built automatically by GitHub Actions when a version tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

[GoReleaser](https://goreleaser.com/) cross-compiles binaries for Linux, macOS, and Windows (amd64 and arm64).

## License

See [LICENSE](LICENSE) for details.
