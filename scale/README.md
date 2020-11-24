# scale

Make your own local environment json and name is env.json.  Use the env.json.example as an example.


Compile
```
go build
```

Add nodes to PCC
```
./scale addNode -n <count>


./scale addNode -n 2
Summary
=======
node  20 192.0.2.10 provision [Ready] connection [online] elapsed 11m19.271488s
node  21 192.0.3.10 provision [Ready] connection [online] elapsed 12m17.186534s
node  22 192.0.2.11 provision [Ready] connection [online] elapsed 12m25.422390s
node  23 192.0.3.11 provision [Ready] connection [online] elapsed 12m21.535521s

Average for 4 nodes - 12m5.853983624s
```

Delete nodes to PCC
```
./scale delNode -n <count>
```

Node Summary (execute same API calls as PCC node summary page)
```
./scale nodeSummary
GetTenants           3 = 690.019324ms
GetNodeGroups        2 = 619.529188ms
GetRoles             6 = 501.196209ms
GetTemplates        14 = 420.923705ms
GetNodesOptions     13 = 4.274975089s

Total elapsed          = 6.506643515s
```
