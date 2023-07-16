---
title: envconfig - read configuration data from environment variables
format: standard
require_prism: true
---

# envconfig - read configuration data from environment variables

[envconfig](https://github.com/vrischmann/envconfig) is a Go library which allows you to define a Go struct representing your configuration object and parsing the configuration data from environment variables.

Every environment variable maps to a field in the configuration struct. It's especially useful when deploying a container because every container runtime supports passing environment variable so you don't need another mechanism to read a configuration file for example.

# Usage

Here is a complete example:

```go
package main

import (
    "fmt"
    "log"

    "github.com/vrischmann/envconfig"
)

type Conf struct {
    MySQL struct {
        Host string
        Port int
        User string
        Password string
    }
    LogPath string
}

func main() {
    var conf Conf
    if err := envconfig.Init(&conf); err != nil {
        log.Fatal(err)
    }

    fmt.Printf("hostname: %s port: %d", conf.MySQL.Host, conf.MySQL.Port)
}
```

Now if you run this like this

```shell
MYSQL_HOST=localhost MYSQL_PORT=3306 go run main.go
```

It will output this

```text
hostname: localhost port: 3306
```

# How it works

I hope how it works is intuitive from the example above: **envconfig** generates a variable name by recursively going through the fields of the root struct you pass to `envconfig.Init` and joins the name with a `_` to generate the final name.

So for example, with this struct:

```go
type Config struct {
    Replicas {
        Slave struct {
            ConnString string
        }
        Master struct {
            ConnString string
        }
    }
}
```

**envconfig** will look for these keys:
* `REPLICAS_SLAVE_CONNSTRING`
* `REPLICAS_MASTER_CONNSTRING`

# Learn more

**envconfig** can do a lot more than that: define optional fields, override the environment variable lookup name, unmarshal using a custom unmarshaler type, etc.

You can learn all of this by looking at the [reference documentation](https://pkg.go.dev/github.com/vrischmann/envconfig).
