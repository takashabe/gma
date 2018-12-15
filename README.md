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

files to use:

- main.go

```go
package main

import (
  "fmt"

  "github.com/takashabe/gma/example/util"
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

---

execute `gma`:

```go
$ gma -main example/main.go -depends example/util.go -depends example/util/util.go
package main

import (
  "fmt"
)

func main() {
  fmt.Println(_util_Foo())
  Foo()
}
func Foo() {
}
func _util_Foo() string {
  return fmt.Sprintf("util")
}
```
