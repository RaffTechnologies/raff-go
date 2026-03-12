# raff-go

Go client library for the [Raff Cloud API](https://rafftechnologies.com).

Used by [raff-cli](https://github.com/rafftechnologies/raff-cli) and [terraform-provider-raff](https://github.com/rafftechnologies/terraform-provider-raff).

## Install

```bash
go get github.com/rafftechnologies/raff-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/rafftechnologies/raff-go"
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
    project, _, err := client.Projects.Create(ctx, &raff.ProjectCreateRequest{
        Name:          "my-project",
        DefaultRegion: "us-east",
    })

    // Create a VM
    vm, _, err := client.VMs.Create(ctx, &raff.VMCreateRequest{
        Name:       "web-01",
        TemplateID: "5ac21891-32e6-41ce-8a93-b5d6ab708b0d",
        PricingID:  3,
        Region:     "us-east",
        ProjectID:  project.ID,
        SSHKeys:    []string{"ssh-rsa AAAA..."},
    })

    // VM actions
    client.VMs.Stop(ctx, vm.ID)
    client.VMs.Start(ctx, vm.ID)
    client.VMs.Reboot(ctx, vm.ID)
}
```

## Custom Configuration

```go
client := raff.New(httpClient, "raff_pub_xxx",
    raff.SetBaseURL("https://api.raffusa.com"),
    raff.SetUserAgent("my-app/1.0"),
    raff.SetProjectID("project-uuid"),
)
```

## License

MIT
