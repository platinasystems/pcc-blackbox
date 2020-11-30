# parseStats

When the scale utility is run on the same CPU as PCC it will create the file
"containers.json" which is a mapping of container ID and container name.  It
will also create a file "container-stats.txt" with the raw container stats. This
utility does post processing off those files.

Compile
```
go build
```

Post process containers.json and container-stats.txt in the current directory.
```
./parseStats
```