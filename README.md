# godin

Godin is an open source monitoring server and agent for linux systems. Its main feature is currently monitoring the state of installed packages and their upgrades.


## Installing

### Server

Godin server requires PostgreSQL database, prefferably version 13 and up.

```sql
create database godin;
create user godin with encrypted password 'password';
grant all privileges on database godin to godin;
```


