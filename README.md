# GOdin

GOdin is an open source monitoring server and agent for linux systems. Its main feature is currently monitoring the state of installed packages.
In current state GOdin can generate and store reports. It is intended to use with visualising software (ex. Grafana).

## How does it work?

GOdin client is a bash script that creates structured JSON report about system state and sends it via curl to the specified server. Report consists of basic host info, installed packages, package upgrades and in case of RPM based distros a section about repositories. The server validates and stores JSON as a report inside PostgresDB using JSONB column. From there visualisation tools are able to query for the data.

A sample Grafana dashboard can be aquired **[here](https://grafana.com/grafana/dashboards/14939)**

## Compatiblity between versions

Current versioning schema assumes compatibility between server and client that match first two decimal points, for example:

```
+----------+-----------+--------------+
|  server  |  client   |  compatible  |
|----------+-----------+--------------|
|  1.0     |  1.0      |     yes      |
|  1.2.1   |  1.2.3    |     yes      |
|  1.3     |  1.2.5    |     no       |
+----------------------+--------------+
```

## Compiling

Only the server requires compiliation, it can be done with:

```
go get
go build
```

## Installing

### Server

Godin server requires PostgreSQL database, prefferably version 13 and up. You can create the database with commands provided below:

```sql
create database godin;
create user godin with encrypted password 'password'; -- REMEMBER TO CHANGE PASSWORD!
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
	"SQLServerAddr": "localhost"
}
```

If you require HTTPS or more advanced setup it is currently suggested you use reverse-proxy software such as Nginx or Apache.

### Client

Place client script in your desired bin directory and call it periodically (using CRON or other mechanisms).

You can modify settings in `/etc/godin/godin-client.conf`. The settings are loaded via sourcing, so do not place whitespace in between `=` signs.

```bash
CLIENT_HOSTNAME="myHostname"
SERVER_URL="http://godin.example.com/reports/upload"
TAGS="tag1,tag2"
```
