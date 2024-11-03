# OnlyConfig

OnlyConfig is a distributed, easy-to-use and powerful configure system.

**Key characteristics**

1. Distributed configure: Supporting as many applications as you want with strong reliability, performance, flexibility.
2. Easy-to-use: Speed up your development from tiny applications to large.
3. Flexible: Custom your own server and client to extend the possibility of configuration.
4. Minimum dependencies: Simplify dependencies, less complexity.
5. Frequently upgrade: supporting more requirements in reality, latest go version with dependencies and modern software
   and hardware

[TOC]

[LOGO]

## 1. Getting started

### 1.1 Server

1. Setup ``postgresql`` database.
2. Create database and import database ddl files in the ``docs`` folder.
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
        SelectorApp:         "app1",
        SelectorEnvironment: "env1",
        SelectorDatacenter:  "dc1",
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

* Start web manager

```text
./webmgr -http=:8880 -postgres=postgres://admin:admin127.0.0.1:5432/onlyconfig
```

* Open browser and add configurations to the app

### 1.4 Getting start and have fun with OnlyConfig !!!

## 2. Features

* [x] OnlyConfig server
* [x] OnlyConfig server: postgresql database storage
* [x] OnlyConfig client: Go language
* [x] OnlyConfig web manager
* [x] OnlyConfig agent
* [ ] OnlyConfig cmd
* [ ] OnlyConfig cleanjob

## 3. User Guide

### 3.1 Server

#### 3.1.1 Start server manually

TBD

#### 3.1.2 Start server using docker

TBD

#### 3.1.3 Built server

```text
go get ./...
go build
```

### 3.2 Go Client

* Suggested usage - using struct as configure container

TBD

* General usage - using general api

TBD

**Note**

1. Configure callback should not raise any panic which will cause client background update task exit

### 3.3 Web Manager

TBD

* Levels of configuration(As selectors)
    * Environment: default envs including DEV, UAT, PreProduction, PROD, etc.
    * Datacenter
    * Namespace(groups of an application)
    * App
* Configure key properties
    * Group
    * Key
    * Value
    * Version(auto generated)
* Configure content type(supporting validation before publishing)
    * General(text)
    * Json
    * Toml
    * Yaml
    * Properties
    * Xml
    * Html

Pending implemented items - TBD

* Link public namespaces to apps as reference configurations
* Release history/Rollback
* Multi-datacenter comparison/publish
* Role based management: user, owner, administrator
* Special handling for production or specific env: reviewing, different appearance
* Better role/naming control: including namespace naming format, etc.
* More configure content editor support
* Support binary file as configure

**Note**

1. Currently, items including application, environment, datacenter, namespace and others could not be deleted due to the
   possibility to cause production issues. If the records are required to be removed(for example offline an
   application or a datacenter), a manual cleanup task is required. Please refer to `cleanjob` tool.

### 3.4 OnlyAgent

OnlyAgent is an independent agent to pull configurations from onlyconfig server and write them to files.

The purpose is to support configuration usage in the applications that don't have available clients in the programming
languages.

#### How to use

##### Method 1: Pull single configuration without extra preparation

```text
./onlyagent -sel dc=dc1,env=DEV -optsel beta=1 -group group1 -key key1 -output application.yaml -hook pwd -server http://127.0.0.1:8800 -server http://127.0.0.2:8800
```

##### Method 2: Pull multiple configuration with configure file provided

Prepare configuration file, for example: `demo.toml`

```text
[[config_list]]
selectors = "dc=dc1,env=DEV"
optional_selectors = "beta=1"
group = "group1"
key = "key1"
output = "application.yaml"
hook = "pwd"

[[config_list]]
selectors = "dc=dc2"
optional_selectors = "beta=2"
group = "group2"
key = "key2"
output = "application2.properties"
```

Start agent

```text
./onlyagent -config demo.toml -server http://127.0.0.1:8800 -server http://127.0.0.2:8800
```

#### Hook

`Hook` is used as a callback when configurations are written to files and ready to load.

`Hook` in the parameter or configure file should be an executable file, including commands or scripts or etc.

When the hook is invoked, the following environment variables are set for the hook to consume:

* ONLYAGENT_GROUP
* ONLYAGENT_KEY
* ONLYAGENT_SEL
* ONLYAGENT_OPTSEL

#### Example usage

```text
Run agent:
go run . -sel app=onlyconfig,dc=default,env=PROD -group GeneralOrg.onlyconfig -key notice -output application.log -hook "D:\hook.bat" -server http://10.11.0.7:8800

Output file:
application.log

Hook file:
D:\hook.bat

In the hook file, the following environment variables are set:
ONLYAGENT_GROUP=GeneralOrg.onlyconfig
ONLYAGENT_KEY=notice
ONLYAGENT_SEL=app=onlyconfig,dc=default,env=PROD
ONLYAGENT_OPTSEL=
```

### 3.5 Cleanjob

TBD

### 3.6 OnlyConfigCommand - ocmd

TBD

## 4. Design

TBD
