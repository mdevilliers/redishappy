

```
go get
```

```
go install
```

```
redishappy
```

To test the json end point -

```
curl -X POST -H "Content-Type: application/json" \
-d '{"method":"HelloService.Say","params":[{"Who":"Test"}], "id":"1"}' \
http://localhost:8085/rpc
```