# raff-go — Public Go Client Library

Shared Go client for the Raff public API. Used by `raff-cli` and `terraform-provider-raff`.

## Sync Rule

**This library MUST stay in sync with the public API spec at `docs/api-reference/openapi.yaml`.**

- Spec adds/changes endpoints → update raff-go to match
- raff-go needs a new function → check spec first, add to spec if missing
- Sync order: **spec → raff-go → raff-cli → terraform-provider-raff**

## Critical Rules

1. **Public API only** — use `docs/api-reference/openapi.yaml` as the sole source of truth. Never reference the internal spec (`raff-api.yaml`).
2. **Never expose admin fields** — no `X-Account-ID`, no admin endpoints, no admin-only response fields.
3. **Auth: `X-API-Key` header only** — account is derived from the key. No account-id parameter needed.
4. **Functions map 1:1 to API endpoints** — each public endpoint gets one Go function.
5. **Changes here affect CLI and Terraform** — test both after any change.

## Structure

```
raff.go       # Client struct, auth, HTTP methods
vms.go        # VM CRUD + actions (start/stop/reboot/resize)
projects.go   # Project CRUD
```

## Quick Commands

```bash
go test ./...    # Run tests
go vet ./...     # Vet
```
