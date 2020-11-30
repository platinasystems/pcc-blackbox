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

Get Events (default page=0, limit=50, search="")
```
./scale$ ./scale getEvent -l 5
5 events found
1   test1vm4        node.avail:info      Node availability: node is back online             11/20 21:57:00 PST
2   test1vm4        system:info          [AGENT][COLLECTOR] have been installed             11/20 21:57:00 PST
3   test2vm10       node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:57:00 PST
4   test2vm10       system:info          [AGENT][COLLECTOR] have been installed             11/20 21:57:00 PST
5   test2vm6        system:info          [AGENT][COLLECTOR] have been installed             11/20 21:57:00 PST
elapsed 641.854321ms
```
```
stig@invader28:~/go/src/github.com/platinasystems/pcc-blackbox6/scale$ ./scale getEvent -l 5 -s "level~warn"
5 events found
1   test1vm2        node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:58:00 PST
2   test2vm10       node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:57:00 PST
3   test2vm9        node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:57:00 PST
4   test1vm7        node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:56:00 PST
5   test1vm10       node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:56:00 PST

elapsed 1.051782232s
```
```
./scale getEvent -l 5 -s "targetName~test1vm1"
5 events found
1   test1vm1        system:info          [AGENT][COLLECTOR] have been installed             11/20 21:54:00 PST
2   test1vm1        node.avail:warn      Node availability: no message received from node for 60 seconds 11/20 21:50:00 PST
3   test1vm1        system:info          [AGENT][COLLECTOR] have been installed             11/20 21:50:00 PST
4   test1vm1        node.avail:info      Node availability: node is back online             11/20 21:49:00 PST
5   test1vm1        system:info          [AGENT][COLLECTOR] have been installed             11/20 21:49:00 PST

elapsed 690.275539ms
```
