---
name: coop-cli
description: Use the coop-cli tool to search for grocery products, manage shopping carts, and find stores on Coop.se (Swedish grocery chain). Activate when the user wants to shop for groceries, search for food products, manage a Coop shopping cart, or find Coop stores.
---

# coop-cli — Coop.se Grocery Shopping CLI

You have access to `coop-cli`, a command-line tool for interacting with Coop.se online grocery shopping.

## Setup

### Authentication

All commands except `stores` require Coop.se credentials. Set them as environment variables:

```bash
export COOP_EMAIL="user@example.com"
export COOP_PASSWORD="password"
```

Or pass them as flags: `--email` / `--password`.

### Store ID

Most commands require a store ID. Set it as an environment variable:

```bash
export COOP_STORE="012345"
```

Or pass it as a flag: `--store`. Find store IDs with the `stores` command.

If credentials or store ID are missing, prompt the user to set them up before proceeding.

## Commands

### Find stores (no auth required)

```bash
coop-cli stores <query>
```

Search by city, name, or address. The `STORE ID` column is the value to use with `--store` or `COOP_STORE`.

### Search for products

```bash
coop-cli search <query> [--limit N]
```

Returns: ID, name, manufacturer, size, price, comparative price, and category. Default limit is 10 results.

### Cart management

```bash
coop-cli cart list                      # Show cart contents
coop-cli cart add <product-id> [qty]    # Add product (default qty: 1)
coop-cli cart remove <product-id>       # Remove product
coop-cli cart clear                     # Empty the cart
```

Product IDs come from `search` results.

### Self-update

```bash
coop-cli update
```

## Workflow

When the user asks to shop for groceries or manage their cart, follow this pattern:

1. If no store is configured, help them find one with `coop-cli stores <city or name>`
2. Search for products with `coop-cli search <query>`
3. Present results and let the user pick which products to add
4. Add selected products with `coop-cli cart add <product-id> [qty]`
5. Show the cart with `coop-cli cart list` to confirm

## Tips

- Product names and search queries are in Swedish (e.g., "mjolk" for milk, "brod" for bread)
- The comparative price (e.g., "kr/kg" or "kr/l") helps compare value across package sizes
- Cart operations return the full updated cart, so there's no need to call `cart list` after `cart add`
