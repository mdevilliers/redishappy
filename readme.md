

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

```
go test ./... -cover
```


To read
-------

https://medium.com/@Drew_Stokes/actual-zero-downtime-with-haproxy-18318578fde6
