# Auth Flutterando with Golang JWT and hasura
---

## Install docker and docker-compose
@todo Consult documentation ...

## Configure your database

For perfect operation, use the following user table as a basis
```sql
CREATE TABLE users (
  id int pk not null unique,
  name text,
  mail text not null unique,
  password text
)


```
If you wanted to change, it would be interesting to read the source code to modify the points you think are best

## Run

```bash
$ docker-compose up -d
```

## Connect with your application

### Routes

 - `/v1/gettoken`
```bash
    $ curl -X POST /auth/v1/gettoken -H 'Authorization: Basic BASE64(USER:PASS)'
```

 - `/v1/check`  
```bash
    $ curl -X POST /auth/v1/check -H 'Authorization: Bearer TOKEN_JWT'
```

 - `/v1/checkMail`  
```bash
    $ curl -X POST /auth/v1/checkMail -H 'Content-Type: application/json' -d '{"name":"João Teste","mail":"joao@test.com", ,"code": 12345678}'
```

 - `/v1/register`  
```bash
    $ curl -X POST /auth/v1/register -H 'Content-Type: application/json' -d '{"name":"João Teste","mail":"joao@test.com", ,"password": "md5-password", "photo": "http://upload-image", "github_user": ""}'
```

 - `/v1/upload` 
```bash
    $ curl -X POST /auth/v1/upload -H 'content-type: multipart/form-data;' -F file=FILE_PATH_ON_CLIENT
```

 - `/v1/uploads/{img}`
```bash
    $ curl -X GET /auth/v1/uploads/FILE_UPLOADED
``` 



