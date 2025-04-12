# go-grep
a golang service that provides API to grep a log file

## How to Run
go run main.go -info-log=info.log -error-log=error.log

### Adding more log files
if you want to add more log files, just add the additional flag: `-log logtype=filepath`

#### Example:
```
go run main.go -info-log=info.log -error-log=error.log -log warning=warning.log -log nginx=nginx.log
```

### Adding Auth
Provide the auth by adding flag: `-auth-header-token={token}`

#### When do request
we need to add an `Authorization` header with the value: `Bearer {token}`

## Future Features
1. Adding more output formats like json format
2. Adding support for multiple servers using peer-to-peer or master node with multiple agents
