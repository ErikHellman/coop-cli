# CLAUDE.md

## Project overview

coop-cli is a Go CLI tool for interacting with Coop.se online grocery shopping. It reverse-engineers the Coop website's APIs (login, product search, cart management, store lookup) and exposes them as simple terminal commands.

## Build and run

```bash
go build ./...          # Build all packages
go run . <command>      # Run directly
go vet ./...            # Lint
```

No external tools or Docker required. Pure Go with no CGO dependencies.

## Project structure

```
main.go                     Entrypoint, sets version info from ldflags
cmd/
  root.go                   Root command, flags, credentials, authenticatedClient()
  search.go                 Product search command
  cart.go                   Cart subcommands (list, add, remove, clear), printCart()
  stores.go                 Store search command (no auth required)
  update.go                 Self-update from GitHub releases
internal/
  auth/auth.go              OIDC login flow against login.coop.se
  api/
    client.go               HTTP client with doRequest/doHybris/doPersonalization helpers
    search.go               Product search via personalization API
    cart.go                 Cart operations via Hybris ecommerce API
    store.go                Store search via store API (standalone, no auth)
  models/models.go          Shared data structures for API responses
```

## Architecture notes

### Authentication flow

The login is a multi-step OIDC flow:
1. GET `www.coop.se/default-login` → HTML form that POSTs to `login.coop.se/connect/authorize`
2. The authorize endpoint redirects to the login page with a `ReturnUrl` query parameter
3. POST credentials to `login.coop.se/local/signin/application-schema/email-password`
4. GET the `ReturnUrl` (authorize callback) → returns a `form_post` HTML page
5. POST the form (code + id_token) to `www.coop.se/signin-oidc` to establish session
6. GET `www.coop.se/api/spa/token` to obtain the Bearer token

The form HTML from login.coop.se uses **single quotes** for attributes (not double quotes). The `parseHiddenForm` regex handles both.

Two HTTP clients share one cookie jar: `noRedirect` (stops on redirects for inspection) and `followRedirect` (normal behavior).

### API layers

Three Coop APIs are used, each with different base URLs and subscription keys:

| API | Base URL | Subscription Key |
|-----|----------|-----------------|
| Hybris ecommerce | `external.api.coop.se/ecommerce/coop` | `3becf0ce306f41a1ae94077c16798187` |
| Personalization (search) | `external.api.coop.se/personalization` | same |
| Store | `proxy.api.coop.se/external/store` | `990520e65cc44eef89e9e9045b57f4e9` |

The Hybris API requires a Bearer token. Store search is unauthenticated. The `doRequest` method in `client.go` is the shared HTTP helper — `doHybris` and `doPersonalization` are thin wrappers that set the right options.

### Cart operations

The `cartdata/.../products` endpoint (used for add/remove) returns the **full cart response**, not a modification status. The `qty=0` convention removes a product.

### Store IDs

The store API returns `storeId` (internal) and `ledgerAccountNumber` (what the ecommerce API uses as store code). The CLI displays `ledgerAccountNumber` as `STORE ID` since that's the value users pass to `--store`.

## Key conventions

- CLI framework: [cobra](https://github.com/spf13/cobra)
- All commands use `RunE` (return errors, don't call `os.Exit`)
- Credentials: `--email`/`--password` flags override `COOP_EMAIL`/`COOP_PASSWORD` env vars
- Store ID: `--store` flag overrides `COOP_STORE` env var; required for search and cart commands (no default)
- Version info is injected via ldflags by goreleaser (`main.version`, `main.commit`, `main.date`)
- No tests exist yet — the APIs are third-party and would require mocking or integration credentials
