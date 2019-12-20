# pcc-blackbox
blackbox go test for pcc

Clone the repository

Make your own local environment json and name is testEnv.json.  Use the testEnv.json.example as an example.

Compile 
```
go test -c
```
There will be a pcc-blackbox.test binary.  Run test
```
./pcc-blabox.test -test.v
```

Options:

\-test.v:  prints test names as test progresses  
\-test.dryrun:  along with \-test.v prints test names but does not execute tests  
\-test.run <name of test group>: excutes just the test group

Available Test Suites:
```
-test.run TestNodes
-test.run TestMaaS
-test.run TestK8s
-test.run TestPortus
```

Example:
```
fyang@i34:~/src/github.com/platinasystems/pcc-blackbox$ ./pcc-blackbox.test -test.v
=== RUN   Test
Iteration 1, Sat Nov 9 23:56:23 2019
=== RUN   Test/nodes
=== RUN   Test/nodes/addNodes
=== RUN   Test/nodes/addNodes/addInvaders
Add id 41 to Nodes
nodeId:41 is  provisionStatus = Adding node...
nodeId:41 is offline provisionStatus = Adding node...
i34 is offline provisionStatus = Adding node...
i34 is offline provisionStatus = Adding node...
i34 is offline provisionStatus = Ready
i34 is offline provisionStatus = Ready
i34 is online provisionStatus = Ready
=== RUN   Test/nodes/delNodes
=== RUN   Test/nodes/delNodes/delAllNodes
i34 provisionStatus = Deleting node...
i34 provisionStatus = Deleting node...
i34 provisionStatus = Deleting node...
i34 provisionStatus = Deleting node...
i34 provisionStatus = Deleting node...
i34 deleted
--- PASS: Test (101.41s)
    --- PASS: Test/nodes (101.41s)
        --- PASS: Test/nodes/addNodes (70.61s)
            --- PASS: Test/nodes/addNodes/addInvaders (70.61s)
        --- PASS: Test/nodes/delNodes (30.80s)
            --- PASS: Test/nodes/delNodes/delAllNodes (30.80s)
PASS
```
