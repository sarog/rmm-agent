### Fork of Tactical RMM Agent
This is a fork of `wh1te909/rmmagent` version `1.5.1` with the build scripts imported from  `1.4.14` before the [source repo](https://github.com/wh1te909/rmmagent) deleted them.

Server source (fork & review coming): https://github.com/wh1te909/tacticalrmm

## Project goals
- Re-introduce an open source version of the agent while maintaining compat with the server.
- Neuter third-party sources or senseless fetching of dependencies online unless absolutely necessary.
- Implement a secure build & delivery system that enforces signature checks.
- Backport changes from the upstream project, if possible.

**The following instructions have not been fully reviewed.**

---

### Building the Windows agent with custom branding

Download and install the following pre-requisites.
- Golang 1.16+
- [Inno Setup](https://jrsoftware.org/isdl.php) (optional)
- Git and your favourite shell.

Clone the repository & download the required dependencies:
```
git clone https://github.com/sarog/rmmagent
go mod download
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

Please review the code & build files and change all references of `Tactical RMM` to `Your Company RMM`

Do __not__ change any of the following or this will break on the RMM end.
- The service names of the 2 Windows services `tacticalagent` and `tacticalrpc`. You can however change the display names and descriptions of these.
- The `TacticalAgent` folder name in Program Files.
- The actual binary name `tacticalrmm.exe`. Change the `FileDescription` in `versioninfo.json` which is what will show up in task manager.

#### Building the 64-bit agent
```
goversioninfo -64
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o tacticalrmm.exe
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build/setup.iss
```

#### Building the 32-bit agent
```
rm resource.syso tacticalrmm.exe
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o tacticalrmm.exe
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build/setup-x86.iss
```

Binaries will be available in ```build\output```

From the RMM, choose the 'Manual' method when generating an agent to get the command line args to pass to the binary.
