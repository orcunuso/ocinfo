# OCinfo

![OCinfo](https://github.com/vorcunus/ocinfo/blob/master/png/ocinfo.png?raw=true)

OCinfo is a tool written in pure Go that was influenced from the hassles of managing multiple OpenShift clusters and the need to improve visibility. What it simple does is to get data from OpenShift APIs with the read-only credentials provided, prints out all the data in a pretty, human-readable and analyzable Microsoft Excel &trade; spreadsheet document and save it as a local xlsx file. With Go, this can be done with an independent binary distribution across all platforms that Go supports, including Linux, MacOS, Windows and ARM.

## Installation

If you have an installed go environment, you can get source code and compile.

```bash
go get github.com/vorcunus/ocinfo
cd $GOPATH/src/github.com/vorcunus/ocinfo
make compile_all
```

If you don't have an installed go environment, you can also download pre-compiled binaries from #TBD.

## Preperation

There are two things that need to be configured or checked before running OCinfo;
