# go-main-aggregator

`go-main-aggregator` provide feature makes an aggregated main file from main and dependencies files.

## Usecase

Most popular use cases is `competitive programming`. Some competitive programming system, require can only submit a soruce code include `main` func. But you can isolation the main and common util files via aggregate `gma`.

## Usage

### Example

```
gma -main main.go -depends util.go -depends util/util.go
```
