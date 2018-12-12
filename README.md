# go-main-aggregator

`go-main-aggregator` provide feature makes an aggregated main file from main and dependencies files.

## Usecase

Most popular use cases are `competitive programming.` Some competitive programming system, require can only submit a source code with include `main` function. You can isolate the main and common utility files via using `gma.`

## Usage

### Example

```
gma -main main.go -depends util.go -depends util/util.go
```
