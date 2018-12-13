# gma (go-main-aggregator)

`gma` provide feature makes an aggregated main file from main and dependencies files.

## Usecase

A most powerful use case is `competitive programming.` Generally, competitive programming should submit to a single file. You can isolate the main(solver) and common utility files via using `gma.`

## Usage

### Example

```
gma -main main.go -depends util.go -depends util/util.go
```
