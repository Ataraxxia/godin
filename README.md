# godin

## Prepare database

Postgres 13 and up

```sql
create database godin;
create user godin with encrypted password 'password';
grant all privileges on database godin to godin;
```
