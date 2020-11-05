# ManagerID

ManagerID is an approach to an ID manager used to identify visitors in our landing pages & services. 

![Passport image](https://i.imgur.com/C1H1E00.jpg)

## Installation

ManagerID needs `go` installed and it uses a MySQL database, that will be auto-initialized the first time that the service runs.

#### 1 - Launch  binary HTTP server
```bash
# You will need the following ENV VARS:

- DB_HOST
- DB_PORT
- DB_USER
- DB_PASS
- DB_NAME

go run main.go
```


##  Usage
You can use ManagerID's API, on the following endpoint:

### `POST` `/id/settle`

```
// Body
{
	"ip": "127.0.8.2",
	"application": "Test Application 2",
	"provider": "Test Provider 2"
}
```
