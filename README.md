# OnlyConfig

OnlyConfig is a distributed, easy-to-use and powerful configure system.

**Key characteristics**

1. Distributed configure: Supporting as many applications as you want with strong reliability, performance, flexibility
2. Easy-to-use: Speed up your development from tiny applications to large.
3. Extensible: Custom your own server and client to extend the possibility of configuration.

[TOC]

## 1. Getting started

### 1.1 Server

1. Setup ``postgresql`` database.
2. Create database and import database ddl in the ``docs`` folder.
3. Start ``OnlyConfig`` server

```text
./onlyconfig -http=:8800 -postgres=postgres://admin:admin127.0.0.1:5432/onlyconfig
```

### 1.2 Go Client

* Suggested usage - using struct as configure container

```go
package main

import (
	"context"

	"github.com/goodplayer/onlyconfig/client"
)

// Define the configure container to use
type ConfigureContainer struct {
	Str  string `json:"str"`
	Int  int    `json:"int"`
	Bool bool   `json:"bool"`
}

func main() {
	// Create the configure client with server list and options
	c := client.NewClient([]string{"http://127.0.0.1:8800"}, client.ClientOptions{
		// Add any selector parameters according to the settings on server side
		SelectorDatacenter: "dc1",
	})
	// Create advanced client
	ca := client.NewClientAdv(c)

	// Register a container with json marshalling and container prototype. Specify the group and key of the configuration.
	// The returned type of container is *atomic.Value . This will guarantee concurrent safe while updating to and reading from the container.
	atomicContainer, _ := ca.RegisterJsonContainer("group_json", "key_json", new(ConfigureContainer))

	// Start the client and register the stop function if needed
	c.StartClient()
	defer c.StopClient()
	// Waiting for all registered configurations loaded for the first time in order for the application to use
	c.WaitStartupConfigureLoaded(context.Background())

	// Use the configuration in the application - retrieving the container everytime
	var container = atomicContainer.Load().(*ConfigureContainer)
}

```

### 1.3 Web Manager

TBD

## 2. Features

* [x] OnlyConfig server
* [x] OnlyConfig server: postgresql database storage
* [x] OnlyConfig client: Go language
* [ ] OnlyConfig web manager

## 3. User Guide

### 3.1 Server Installation

#### 3.1.1 Start server manually

TBD

#### 3.1.2 Start server using docker

TBD

### 3.2 Go Client

* Suggested usage - using struct as configure container

TBD

* General usage - using general api

TBD

### 3.3 Web Manager

## 4. Design

TBD
