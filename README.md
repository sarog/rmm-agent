### Fork of Tactical RMM Agent
This is a fork of `wh1te909/rmmagent` version `1.5.1` with the build scripts imported from `1.4.14` before the [source repo](https://github.com/wh1te909/rmmagent) [deleted them](https://github.com/wh1te909/rmmagent/commit/3fdb2e8c4833e5310840ca79bf394a53f6dbe990).

Server source (fork & review coming): https://github.com/wh1te909/tacticalrmm

## Project goals
- Re-introduce an open source version of the agent while maintaining compat with the server.
- Make the Python dependency optional.
- Implement a secure build & delivery system that enforces signature checks, or allows the sysadmin/developer to sign their own builds.
- Backport changes from the upstream project, if possible.

**The following instructions have not been fully reviewed.**

---

### Building the Windows agent with custom branding

Pre-requisites:
- Golang 1.16+
- [Inno Setup](https://jrsoftware.org/isdl.php) (optional) for distribution

Clone the repository & download the dependencies:
```
git clone https://github.com/sarog/rmmagent
go mod download
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

Please review the code & build files: change all references of `Tactical RMM` to `Your Company RMM`

Do __not__ change any of the following or this will break on the RMM end.
- The service names of the two (2) Windows services `tacticalagent` and `tacticalrpc`. However, the service display name & description **can** be changed.
- The `TacticalAgent` folder name in Program Files.
- The actual binary name `tacticalrmm.exe`. Change the `FileDescription` in `versioninfo.json` which is what will show up in Task Manager.

#### Building the 64-bit agent
```
goversioninfo -64
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o out\tacticalrmm.exe
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup.iss
```

#### Building the 32-bit agent
```
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o out\tacticalrmm.exe
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup-x86.iss
```


From the RMM, choose the 'Manual' method when generating an agent to get the command line args to pass to the binary.
