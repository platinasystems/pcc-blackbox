# pcc-blackbox
blackbox go test for pcc

**Clone the repository**

Make your own local environment json and name is testEnv.json.  Use the testEnv.json.example as an example.

**Compile** 
```
go test -c
```
There will be a pcc-blackbox.test binary.  

**Execution**

To start the BB, give the following command
```
./pcc-blackbox.test -test.v 
```

Options:

\-test.v:  prints test names as test progresses  
\-test.dryrun:  along with \-test.v prints test names but does not execute tests  
\-test.run <name of test group>: executes just the tests in the specified group (see the list of available test groups below)  
\-test.run TestCustom [--testfile=file_name_here]: execute the tests specified in the file testList.yml (see the example below). A specific file name could be given using the option --testfile  

List of available Test Groups:
```
  TestNode 
  TestClean 
  TestNetCluster 
  TestConfigNetworkInterfaces 
  TestMaaS 
  TestUsers 
  TestTenantMaaS 
  TestGreenfield 
  TestHardwareInventory 
  TestAddK8s 
  TestK8sApp 
  TestDeleteK8s 
  TestCeph 
  TestCephCache 
  TestAppCredentials 
  TestPortus 
  TestDelPortus 
  TestUserManagement 
  TestKeyManager 
  TestPolicy 
  TestDashboard 
```

Example of TestCustom file
```
tests:
  TestNodes: []
  TestNode:
    - getNodes
    - updateIpam
    - addNetCluster
```
Note: TestNodes: [] is used to include a Test Group already available.


**Execution Examples:**  
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

**Output & Logs**  

It is possible to configure the output (STDOUT) and log verbosity from the logConfig.yaml file.  
Here is the content of the config file:
 
```
logs:
  appender:
    stdout:
      enabled: true
      level: TRACE
    default:
      enabled: true
      level: INFO
      maxfilesize: 25MB
      maxhistory: 1
      totalsizecap: 400MB
    detailed:
      enabled: true
      level: TRACE
      maxfilesize: 25MB
      maxhistory: 1
      totalsizecap: 400MB
    error:
      enabled: true
      level: ERROR
      maxfilesize: 25MB
      maxhistory: 1
      totalsizecap: 400MB
```
Set **enabled: true** to get the tests output on the STDOUT, **false** otherwise.  
Set level to ERROR or WARN to get less details, INFO, TRACE or DEBUG otherwise.  
The default.log, detailed.log and error.log are in the /logs folder.


**Results rewiew**  
The results for each execution of the BB are hosted in a sqlite database.   
To login into the DB, do:
```
sqlite3 blackbox.db
```
The list of tables in the DB can be obtained with the following command:
```
sqlite> .tables
random_seeds  test_results
```
The table **random_seeds** contains the association between the specific execution and teh related "seed" used in all the random operations.  
The table **test_results** contains the results for each test.  
To get the content of the above tables, do:
```
sqlite> select * from test_results;
1|5e2e5cac-fc6f-4ba8-bc93-ac35421b2284|getAvailableNodes|pass||0.855966701
2|5e2e5cac-fc6f-4ba8-bc93-ac35421b2284|testNodeGroups|pass||2.519426553
3|5e2e5cac-fc6f-4ba8-bc93-ac35421b2284|getSecKeys|pass||0.575807297
4|5e2e5cac-fc6f-4ba8-bc93-ac35421b2284|updateSecurityKey_MaaS|pass||2.874586958
5|5e2e5cac-fc6f-4ba8-bc93-ac35421b2284|addNodesAndCheckStatus|pass||0.000329746
```
Each row contains:

\- an incremental ID,  
\- an UUID representing a specific execution of the BB; e.g all teh test having 5e2e5cac-fc6f-4ba8-bc93-ac35421b2284 have been executed in the same run;  
\- the name of the test;  
\- the result of the test (pass, failed, sipped, undefined);  
\- a message, in case of any error;  
\- the time needed to execute the test;  

Exit from DB with:
```
sqlite> .quit
```


Queries examples  

See how many runs have been execute and how many test for runs  
select run_id, count(*) as c FROM test_results GROUP BY run_id;
```
005d82c3-0dcb-46ca-86ce-d812328f0e5d|20
04a33aeb-26a5-422f-a76c-2683349111f8|9
04e30125-fd6b-4859-83fb-4791ca6f91ae|12
093c6984-f188-44a5-bf19-01a8ca3de044|12
1057c26c-d8f0-47db-97a2-5a0b4e7a5bf9|6
18b20f5c-7f20-4e5b-afa6-a785e2410b5b|10
```
See how many pass, skip, fail, undefined for each run  
select run_id, count(CASE WHEN result='pass' THEN 1 END) as ok, count(CASE WHEN result='fail' THEN 1 END) as nok, count(CASE WHEN result='skip' THEN 1 END) as sk, count(CASE WHEN result='undefined' THEN 1 END) as ud FROM test_results group by run_id;
```
005d82c3-0dcb-46ca-86ce-d812328f0e5d|13|5|0|2
04a33aeb-26a5-422f-a76c-2683349111f8|6|2|0|1
04e30125-fd6b-4859-83fb-4791ca6f91ae|7|0|0|0
093c6984-f188-44a5-bf19-01a8ca3de044|7|0|0|0
1057c26c-d8f0-47db-97a2-5a0b4e7a5bf9|4|1|0|1
18b20f5c-7f20-4e5b-afa6-a785e2410b5b|6|3|0|1
1cba6b7c-cb80-4a9e-90c5-4b5e77932e16|4|0|0|0
21b170a6-245b-4b30-9485-c09ad57d15bc|0|2|0|0
2a926615-7db3-4189-8e4e-6129ca5668b5|4|1|0|1
```
See how many time a specific test has been executed and the average, max and min execution time  
select test_id, count(test_id), avg(elapsed_time)as ta, max(elapsed_time)as tmax, min(elapsed_time)as tmin FROM test_results GROUP BY test_id ORDER BY ta DESC;
testVerifyCephInstallation|1|693337975393.0|693337975393|693337975393
```
validateK8sCluster|5|351144951905.8|1086943860301|542516347
testVerifyCephFSCreation|1|214784112749.0|214784112749|214784112749
testDeleteCeph|1|211706081406.0|211706081406|211706081406
testCeph|6|171221407683.5|1015421423243|1557950414
checkPortus|1|140377444425.0|140377444425|140377444425
testVerifyK8sAppDeployment|1|131162481384.0|131162481384|131162481384
testVerifyK8sAppUnDeployment|1|118508725207.0|118508725207|118508725207
```

**Test template**  
Use the following template to write a new test.

```
func name_of_test_here(t *testing.T, ...) {
	
	//To support the dryrun mode
	test.SkipIfDryRun(t)
    
    //To create and init the test result structure. This guarantee that the result of the, 
    //even if in case of failure, will be correctly store in teh DB at the end of the execution
	res := models.InitTestResult(runID)
	defer res.CheckTestAndSave(t, time.Now())

	//PUT HERE THE BODY OF THE TEST
	//Please, USE the AuctaLogger to be compliant with the "Output & Log management".
    log.AuctaLogger.Info(fmt.Sprintf("\nDoing some interesting stuff))
    
    //MANAGE FAILURE: please, set the ERROR MESSAGE for the DB and logs 
    //before to get the FailNow()
	if err != nil {
		log.AuctaLogger.Error(fmt.Sprintf("Here is the error: %v", err))
		res.SetTestFailure(err.Error())
		log.AuctaLogger.Error(err.Error())
		t.FailNow()
		return
	}
}
```
