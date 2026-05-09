# raff-go

[![CI](https://github.com/RaffTechnologies/raff-go/actions/workflows/ci.yml/badge.svg)](https://github.com/RaffTechnologies/raff-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/rafftechnologies/raff-go.svg)](https://pkg.go.dev/github.com/rafftechnologies/raff-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Official Go client library for the [Raff Cloud API](https://docs.rafftechnologies.com).

Used by [raff-cli](https://github.com/RaffTechnologies/raff-cli) and [terraform-provider-raff](https://github.com/RaffTechnologies/terraform-provider-raff). The low-level HTTP client is generated from the public OpenAPI spec via [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen); the hand-written service wrappers in this package give you a stable Go-idiomatic API.

## Install

```bash
go get github.com/rafftechnologies/raff-go
```

Requires Go 1.25+.

## Authentication

All requests authenticate via an API key (`X-API-Key` header). Generate one in the dashboard at https://rafftechnologies.com under **Team & Projects → API Keys**.

```go
client := raff.NewFromToken("raff_pub_xxx")
```

Or use a custom HTTP client and options:

```go
client := raff.New(
    &http.Client{Timeout: 30 * time.Second},
    "raff_pub_xxx",
    raff.SetBaseURL("https://api.rafftechnologies.com"),
    raff.SetUserAgent("my-app/1.0"),
    raff.SetProjectID("project-uuid"), // sets the X-Project-ID header for project-scoped calls
)
```

## Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    raff "github.com/rafftechnologies/raff-go"
    "github.com/rafftechnologies/raff-go/spec"
)

func main() {
    client := raff.NewFromToken("raff_pub_xxx")
    ctx := context.Background()

    // List projects
    projects, _, err := client.Projects.List(ctx, nil)
    if err != nil {
        panic(err)
    }
    for _, p := range projects {
        fmt.Printf("%s  %s\n", p.ID, p.Name)
    }

    // Create a project
    region := spec.CreateProjectRequestDefaultRegion("us-east")
    project, _, err := client.Projects.Create(ctx, &raff.CreateProjectRequest{
        Name:          "my-project",
        Description:   raff.String("Production workloads"),
        DefaultRegion: &region,
    })
    if err != nil {
        panic(err)
    }

    // Project-scoped calls (creating VMs, VPCs, IPs, etc.) need the
    // X-Project-ID header. Construct a project-scoped client:
    pc := raff.NewFromToken("raff_pub_xxx", raff.SetProjectID(project.ID.String()))

    // Create a VM
    templateID, _ := uuid.Parse("5ac21891-32e6-41ce-8a93-b5d6ab708b0d")
    sshKeys := []string{"ssh-ed25519 AAAA... user@host"}
    vm, _, err := pc.VMs.Create(ctx, &raff.CreateVMRequest{
        Name:       "web-01",
        TemplateID: templateID,
        PricingID:  3,
        Region:     spec.CreateVMRequestRegion("us-east"),
        SSHKeys:    &sshKeys,
    })
    if err != nil {
        panic(err)
    }

    // Power actions
    pc.VMs.Stop(ctx, vm.ID.String())
    pc.VMs.Start(ctx, vm.ID.String())
    pc.VMs.Reboot(ctx, vm.ID.String())
}
```

See [pkg.go.dev](https://pkg.go.dev/github.com/rafftechnologies/raff-go) for the full API reference.

## Services

Twelve services on the client, ~85 operations total:

| Service | Operations |
|---------|------------|
| **Compute** | |
| `client.VMs` | Full lifecycle (29 ops): list, create, delete, start/stop/reboot, resize, rename, reinstall, factory-reset, save-image, attach/detach VPC/IP/security-group, tags, notes |
| **Networking** | |
| `client.VPCs` | List, Get, GetDetail, Create, Update, Delete, CIDRSuggestions |
| `client.IPs` | List, Get, Reserve, Release, Change |
| `client.SecurityGroups` | List, Templates, Get, Create, Update, Delete |
| **Identity & access** | |
| `client.Projects` | List, Get, Create, Update, Delete |
| `client.ProjectMembers` | List, Get, Add, Update, Remove (per-project) |
| `client.Members` | List, Get, Add, Update, Remove (account-level) |
| `client.Roles` | List, Get, Create, Update, Delete |
| `client.Permissions` | List (read-only catalog) |
| `client.Invitations` | CreateAccount, CreateProject, Cancel |
| `client.APIKeys` | List, Get, Create, Update, Regenerate, Revoke |
| `client.SSHKeys` | List, Get, Create, Update, Delete |

## Versioning

This library follows [Semantic Versioning](https://semver.org/). v0.x is allowed to introduce breaking changes; v1.0.0 onward implies a stable public API.

Pin a specific version:

```bash
go get github.com/rafftechnologies/raff-go@v0.2.0
```

The generated client (`spec/spec.gen.go`) is auto-synced with the public OpenAPI spec at [docs/api-reference/openapi.yaml](https://github.com/RaffTechnologies/docs/blob/main/api-reference/openapi.yaml) on a nightly schedule. Spec changes typically result in a PR within 24 hours.

## Documentation

- **API reference** — [docs.rafftechnologies.com](https://docs.rafftechnologies.com)
- **Go SDK reference** — [pkg.go.dev/github.com/rafftechnologies/raff-go](https://pkg.go.dev/github.com/rafftechnologies/raff-go)
- **Dashboard** — [rafftechnologies.com](https://rafftechnologies.com)
- **Public OpenAPI spec** — [github.com/RaffTechnologies/docs/blob/main/api-reference/openapi.yaml](https://github.com/RaffTechnologies/docs/blob/main/api-reference/openapi.yaml)

## Related projects

- [raff-cli](https://github.com/RaffTechnologies/raff-cli) — official command-line interface, built on raff-go
- [terraform-provider-raff](https://github.com/RaffTechnologies/terraform-provider-raff) — official Terraform provider, built on raff-go

## Contributing

PRs welcome. To regenerate the client after a spec change:

```bash
make generate
```

`make verify` enforces drift-free `spec.gen.go` in CI.

## License

[MIT](LICENSE)
