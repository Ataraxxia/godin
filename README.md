# GOdin

GOdin is an open source monitoring server and agent for linux systems. Its main feature is currently monitoring the state of installed packages and their upgrades.
In current state GOdin can generate and store reports sent from its client. It is intended to use with visualising software (ex. Grafana) to visualize gathered data.

## How does it work?

GOdin client is a bash script that creates structured JSON report about packages state and sends it via curl to the specified server. Report consists of basic host info, installed packages, package upgrades and in case of RPM based distros a section about repositories. The server validates and stores JSON as a report inside PostgresDB using JSONB column. From there visualisation tools are able to query for the data.

## Compiling

Only the server requires compiliation, it can be done with:

```
go get
go build
```

## Installing

### Server

Godin server requires PostgreSQL database, prefferably version 13 and up.

```sql
create database godin;
create user godin with encrypted password 'password';
grant all privileges on database godin to godin;
```

Create `/etc/godin/settings.json` file and fill it according to your needs. Make it so same user that runs executable can access this file.

```json
{
        "Address": "0.0.0.0",
        "Port": "80",
        "LogLevel": "INFO",
	"SQLUser": "godin",
	"SQLPassword": "password",
	"SQLDatabaseName": "godin",
	"SQLServerAddress": "localhost"
}
```

If you require HTTPS or more advanced setup it is currently suggested you use reverse-proxy software such as Nginx or Apache.

### Client

Place client script in your desired bin directory and call it periodically using for example CRON.
