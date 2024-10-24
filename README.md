# RMM Agent

A generic Remote Monitoring & Management agent written in Go for sysadmins, by sysadmins.

Downloadable Windows binaries will not be provided on this GitHub repository as they will be unusable due to false-positives from antivirus software. Users are encouraged to [build](#building-the-agent) and [sign](CODESIGN.md) their own executables to guarantee integrity.

## Project goals
- Allow anyone to use and modify the agent for whatever they see fit.
  - This includes using any other future open source RMM server/backend.
- Implement support for additional operating systems, such as FreeBSD, Linux, macOS, and others.

## Building the agent

Pre-requisites:
- [Go](https://go.dev/dl/) 1.21+

### Windows

Clone the repository & download the dependencies. GoVersionInfo is used to optionally generate the Windows file properties and icons.
```shell
git clone https://github.com/jetrmm/rmm-agent
go mod download
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

Building the x64 agent:
```shell
goversioninfo -64
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o out\rmmagent.exe
```

Building the x86 agent:
```shell
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o out\rmmagent.exe
```

### Building the installer
 
Creating an optional installer (setup) file requires [Inno Setup](https://jrsoftware.org/isdl.php) 6.2+ for packaging & distributing the agent

Packaging the x64 installer:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup.iss
```

Packaging the x86 installer:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup-x86.iss
```

## Signing the agent and installer

See [CODESIGN](CODESIGN.md) for more information.

## Branding the agent

Take a look at `agent/const.go` and the defined constants at the top of every file to change the strings.

## Attribution

The JetRMM agent is a hard-fork of [wh1te909/rmmagent](https://github.com/wh1te909/rmmagent) version `1.5.1` (specifically [3b070c3fadb1ae4dff8f20196fc721d99a760474](https://github.com/wh1te909/rmmagent/tree/3b070c3fadb1ae4dff8f20196fc721d99a760474)) which is considered the **last MIT-licensed release** before Amidaware's introduction of the _Tactical RMM License_. Additionally, code from versions 1.5.3 to 1.8.0 (all under the MIT License) may be referenced through reverse engineering.
