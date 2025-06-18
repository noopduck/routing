# Routing Library

## Description

The Routing Library is a Go package designed to interact with the Linux routing table. It provides functionality to retrieve the default gateway address by reading the `/proc/net/route` file, converting hexadecimal values to decimal, and finally to a human-readable IP address format.

## Installation

To use the Routing Library, simply import it into your Go project:

```go
import "github.com/noopduck/routing"
```

```bash
go get github.com/noopduck/routing@latest
```

## Usage

Here is a quick example of how to use the library to find the default gateway:

```go
package main

import (
    "fmt"
    "log"
    "github.com/noopduck/routing"
)

func main() {
    gateway, err := routing.FindLinuxDefaultGW()
    if err != nil {
        log.Fatalf("Error finding default gateway: %v", err)
    }
    fmt.Printf("Default Gateway: %s\n", gateway)
}
```

```go
package main

import (
    "fmt"
    "log"
    "github.com/noopduck/routing"
)

func main() {
    routes, err := routing.GetLinuxRoutingTable()
    if err != nil {
        log.Fatalf("Error retrieving routing table: %v", err)
    }
    for _, route := range routes {
        fmt.Printf("Interface: %s, Destination: %s, Gateway: %s\n", route.Interface, route.Destination, route.Gateway)
    }
}
```
## License

This project is licensed under the MIT License - see the LICENSE file for details.
