# Financify

## Generating a private RSA key using `openssl` tool

```bash
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private.pem -out public.pem
```

## Commands

### Monitoring command

```expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,mem:memstats.Alloc"```

### Get token

```curl --user "admin@example.com:gophers" http://localhost:3000/v1/token/90a50c59-e095-4c36-b9a3-54f83a3832e2```

### Load the service command

```hey -m GET -c 100 -n 10000 -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1/users```
