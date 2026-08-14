# go-pure-flasharray

!! Message from maintainer: This project is currently under development. Don't use it on any production system. I'm currently looking for a test environment to properly validate the code. Once past that, I will delete this message and release v1.0.0. 

## About

A Go client library for the Pure Storage FlashArray REST API v2, along with an HTTP mock server for testing.

## Packages

| Package | Description |
|---------|-------------|
| `pkg/flashclient` | FlashArray REST API client |
| `pkg/mock` | In-process mock server for use in tests |
| `cmd/mock` | Standalone mock server binary |

## Installation

```bash
go get github.com/sanderdescamps/go-pure-flasharray
```

## API Client

### Creating a client

```go
import "github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"

client, err := flashclient.NewRestClient(
    "https://my-flasharray.example.com",
    "my-api-token",
)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

The client automatically negotiates the latest supported API version. To pin a specific version or configure other options, pass a `ClientConfig`:

```go
cfg := flashclient.ClientConfig{
    ApiVersion: "2.40",
    UserAgent:  "my-app/1.0",
    Insecure:   true,  // skip TLS verification
    Debug:      false,
}

client, err := flashclient.NewRestClient("https://my-flasharray.example.com", "my-api-token", cfg)
```

### Supported resources

The client provides methods for the following resources:

| Resource | Key methods |
|----------|-------------|
| Arrays | `GetArrays()` |
| Volumes | `GetVolumes()`, `GetVolume(id)`, `GetVolumeByName(name)`, `CreateVolume(name, post)`, `UpdateVolume(id, patch)`, `DeleteVolume(id)` |
| Hosts | `GetHosts()`, `GetHost(id)`, `GetHostByName(name)`, `CreateHost(name, post)`, `UpdateHost(id, patch)`, `DeleteHost(id)` |
| Host Groups | `GetHostGroups()`, `GetHostGroupByName(name)`, `CreateHostGroup(name)`, `UpdateHostGroup(id, patch)`, `DeleteHostGroup(id)` |
| Connections | `GetConnections()`, `CreateHostConnections(hostNames, volumeIds, post)`, `CreateHostGroupConnections(hostGroupNames, volumeIds, post)`, `DeleteConnections(hostNames, volumeIds)` |
| Pods | `GetPods()`, `GetPod(id)`, `GetPodByName(name)`, `CreatePod(name, post)`, `UpdatePod(id, patch)`, `DeletePod(id)` |
| Volume Groups | `GetVolumeGroups()`, `GetVolumeGroup(id)` |
| Volume Snapshots | `GetVolumeSnapshots()`, `GetVolumeSnapshot(id)` |
| Protection Groups | `GetProtectionGroups()`, `GetProtectionGroupByName(name)` |
| Alerts | `GetAlerts()` |
| Controllers | `GetControllers()` |
| Drives | `GetDrives()` |
| Hardware | `GetHardware()` |
| Network Interfaces | `GetNetworkInterfaces()` |
| Ports | `GetPorts()` |

### Example: listing and creating volumes

```go
// List all volumes
volumes, err := client.GetVolumes()
if err != nil {
    log.Fatal(err)
}
for _, v := range volumes {
    fmt.Printf("Volume: %s  ID: %s  Size: %d bytes\n", v.Name, v.Id, v.Provisioned)
}

// Create a 100 GiB volume
provisioned := int64(100 * 1024 * 1024 * 1024)
vol, err := client.CreateVolume("my-volume", flashclient.VolumePost{
    Provisioned: provisioned,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created volume: %s (%s)\n", vol.Name, vol.Id)
```

### Example: connecting a host to a volume

```go
// Create a host with FC WWNs
host, err := client.CreateHost("my-host", flashclient.HostPostBody{
    WWNs: []string{"21:00:00:24:ff:aa:bb:cc"},
})
if err != nil {
    log.Fatal(err)
}

// Connect the host to the volume
connections, err := client.CreateHostConnections(
    []string{host.Name},
    []string{vol.Id},
    flashclient.ConnectionPostBody{},
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Connected host %s to volume %s (LUN %d)\n",
    connections[0].Host.Name, connections[0].Volume.Name, *connections[0].Lun)
```

## Mock Server

The mock server implements the FlashArray REST API in memory, making it suitable for integration tests without a real array.

### Using the mock in tests

```go
import (
    "testing"
    "github.com/sanderdescamps/go-pure-flasharray/internal/fakearray"
    "github.com/sanderdescamps/go-pure-flasharray/pkg/flashclient"
    "github.com/sanderdescamps/go-pure-flasharray/pkg/mock"
)

func TestMyCode(t *testing.T) {
    // Start a mock server with embedded test data
    array, err := fakearray.NewArrayWithTestData()
    if err != nil {
        t.Fatal(err)
    }

    server := mock.NewMock(array)
    go server.Start("127.0.0.1", 8080)
    defer server.Stop()

    // Connect the client to the mock
    client, err := flashclient.NewRestClient("http://127.0.0.1:8080", "fake-auth-token")
    if err != nil {
        t.Fatal(err)
    }
    defer client.Close()

    volumes, err := client.GetVolumes()
    // ...
}
```

To start with an empty array instead:

```go
array := fakearray.NewArray()
server := mock.NewMock(array)
```

To load data from a directory exported from a real array:

```go
array, err := fakearray.NewArrayFromDir("/path/to/export")
```

### Running the mock server binary

Build and run the standalone binary:

```bash
go build -o go-purefa-mock ./cmd/mock

# Start with embedded test data (no auth)
./go-purefa-mock run --test-data --disable-auth

# Start with data exported from a real array
./go-purefa-mock run --data-dir ./tests/gi-export-san20010 --disable-auth

# Start with custom credentials on a specific address
./go-purefa-mock run --test-data --username admin --api-token my-token --address 0.0.0.0 --port 9090
```

| Flag | Default | Description |
|------|---------|-------------|
| `--address`, `-a` | `127.0.0.1` | Listen address |
| `--port`, `-p` | `8080` | Listen port |
| `--username`, `-u` | `pureuser` | Auth username |
| `--api-token`, `-t` | `fake-auth-token` | Auth API token |
| `--test-data`, `-e` | false | Use embedded test data |
| `--data-dir`, `-d` | — | Load test data from directory (mutually exclusive with `--test-data`) |
| `--disable-auth`, `-s` | false | Disable authentication checks |
| `--log`, `-l` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

### Exporting data from a real array

The `export` command connects to a live FlashArray and dumps its state to a directory. The resulting directory can be used as `--data-dir` input for the mock server.

```bash
./go-purefa-mock export \
    --endpoint https://my-flasharray.example.com \
    --api-token my-api-token \
    --export-dir ./my-array-export \
    --insecure
```

## Running tests

Tests use the mock server automatically when no real array environment variables are set.

```bash
go test ./...
```

To run tests against a real FlashArray, set the following environment variables:

| Variable | Description |
|----------|-------------|
| `PUREFA_TEST_ENDPOINT` | Array endpoint URL |
| `PUREFA_TEST_API_TOKEN` | API token |
| `PUREFA_TEST_INSECURE` | Set to `true` to skip TLS verification |
| `PUREFA_TEST_API_VERSION` | Pin a specific API version |
| `PUREFA_TEST_DEBUG` | Set to `true` to enable request debug logging |

## License

See [LICENSE](LICENSE).
