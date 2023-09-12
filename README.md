# RMM Agent

A generic Remote Monitoring & Management agent written in Go.

This is a fork of [wh1te909/rmmagent](https://github.com/wh1te909/rmmagent) version `1.5.1`, the last MIT-licensed release before Amidaware's introduction of the _Tactical RMM License_.

**Please note**: downloadable binaries (executables) will not be provided on this GitHub repository as they will be useless. Users are encouraged to [build](#building-the-windows-agent) and [sign](CODESIGN.md) their own executables to guarantee integrity.

## Project goals
- Allow anyone to use and modify the agent for whatever they see fit.
  - This includes using any other future open source RMM server/backend.
- Implement support for additional operating systems, such as FreeBSD, Linux, macOS, and others.

## Building the agent

Pre-requisites:
- [Go](https://go.dev/dl/) 1.21+

### Windows

Clone the repository & download the dependencies. GoVersionInfo is used to generate the Windows file properties and icons.
```
git clone https://github.com/sarog/rmm-agent
go mod download
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

Building the x64 agent:
```
goversioninfo -64
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o out\rmmagent.exe
```

Building the x86 agent:
```
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o out\rmmagent.exe
```

## Signing the agent

See [CODESIGN](CODESIGN.md) for more information.

## Building the installer
 
Creating an optional installer (setup) file requires [Inno Setup](https://jrsoftware.org/isdl.php) 6.2+ for packaging & distributing the agent

Packaging the x64 installer:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup.iss
```

Packaging the x86 installer:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup-x86.iss
```

## Updating the agent

This is currently unimplemented; auto-updating agents will be worked on at a later date.

## Branding the agent

Take a look at the constants at the top of every file to change the displayed names.
