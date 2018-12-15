# gma (go-main-aggregator)

`gma` provide feature makes an aggregated main file from main and dependencies files.

## Usecase

A most powerful use case is `competitive programming.` Generally, competitive programming should submit to a single file. You can isolate the main(solver) and common utility files via using `gma.`

## Installation

```
go get github.com/takashabe/gma
```

_require go1.11 or later and `GO111MODULE` environemnts_


## Usage

```
gma -main main.go -depends util.go,util2.go -depends util/util.go
```

#### Options

```
-main    require: the central file of aggregation
-depends optional: dependencies files
```

### Example

- main.go

```go
package main

import (
  "fmt"

  "github.com/takashabe/go-main-aggregator/example/util"
)

func main() {
  fmt.Println(util.Name)
  fmt.Println(util.Foo())
  Foo()
}
```

- util.go

```go
package main

func Foo() {}
```

- util/util.go

```go
package util

import "fmt"

const Name = "util.go"

func Foo() string {
  return fmt.Sprintf("util")
}
```

--

```go
$ gma -main main.go -depends util.go,util2.go -depends util/util.go > tmp
package main

import (
        "fmt"
        "github.com/takashabe/gma/example/util"
)

func main() {
        fmt.Println(util.Name)
        fmt.Println(_util_Foo())
        Foo()
}
func Foo() {
}

const Name = "util.go"

func _util_Foo() string {
        return fmt.Sprintf("util")
}
```
