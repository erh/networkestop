# networkestop
network checking estop module for viam

sample config
```
{
   "server" : "8.8.8.8",
   "lookup" : "www.viam.com.",
   "stop" :   ["myBase"]
}
```

to compile for arm64
====
```
env GOOS=linux GOARCH=arm64 make
```
