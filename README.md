# go-grep
a golang service that provide API to grep a log file

## Example How to Run
go run main.go -info-log=info.log -error-log=error.log

### Adding more log files
if you want to add more log files, just adding the additional flag: -log logtype=filepath

#### Example:
go run main.go -info-log=info.log -error-log=error.log -log warning=warning.log -log nginx=nginx.log

### Adding Auth
provide the auth by adding flag: -auth-header-token={token}

#### When do request
we need to add "Authorization" header with value: "Bearer {token}"