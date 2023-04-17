# RMM Agent

This is a fork of [wh1te909/rmmagent](https://github.com/wh1te909/rmmagent) version `1.5.1` with the build scripts imported from `1.4.14` before the source repo [deleted them](https://github.com/wh1te909/rmmagent/commit/3fdb2e8c4833e5310840ca79bf394a53f6dbe990). It is considered the last MIT-licensed release before Amidaware's introduction of the _Tactical RMM License_.

This `dev` branch is focusing on making the agent on par with 1.7.2 features from upstream. It is considered incomplete and unfit for production use, but feel free to test.

**Please note**: downloadable binaries (executables) will not be provided on this GitHub repository as they will be useless. Users are encouraged to [build](#building-the-windows-agent) and [sign](CODESIGN.md) their own executables to guarantee integrity.

## Project goals
- ~~Re-introduce an open source version of the agent.~~ ✅
- Allow anyone to use and modify the agent for whatever they see fit.
  - This includes using any other future open source RMM server/backend.
- ~~Make the Python dependency optional.~~ ✅
- ~~De-brand the agent.~~ ✅

### Differences between upstream and this repo

- The Python binaries are no longer downloaded by default.
  - If Python is desired, it will be possible to install it through [Chocolatey](https://community.chocolatey.org/packages/python) manually.
  - (todo) Chocolatey will be used to (optionally) install Python in an upcoming release.
  - As a result, new PowerShell commands have replaced the Python scripts in functions `LoggedOnUser()` and `GetCPULoadAvg()`.
- New CLI flag `-windef` allows control over adding Windows Defender exclusions.
  - The default value is `false`, meaning the agent executable & related paths will not be added to Windows Defender's exclusions.
- New CLI flag `-py-enabled` allows control over whether Python installation or execution is allowed.
  - A new Registry value called `PythonEnabled` is created.
  - The default value is `false`, meaning Python will not be downloaded, and `.py` scripts will not execute.
- (todo) Automatic updates from the RMM server will be temporarily disabled for obvious reasons.
- This agent is **unofficial & unsupported**, so don't bother the upstream developers about it (but feel free to [create an issue](https://github.com/sarog/rmmagent/issues/new) here).

### What's missing

As of `2022-02-22`:
- The CLI flags ~~-nomesh and~~ `-meshdir` have not been implemented.
- A few of the Task Scheduler / Automated Tasks functionality is (probably) incomplete.
- Features from agent version 1.8.0 have not been addressed yet.

### Building the Windows agent

Pre-requisites:
- [Go](https://go.dev/dl/) 1.19+
- [Inno Setup](https://jrsoftware.org/isdl.php) 6.2+ for (optionally) packaging & distributing the agent

Clone the repository & download the dependencies:
```
git clone https://github.com/sarog/rmmagent
go mod download
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

#### Building the 64-bit agent
```
goversioninfo -64
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o out\rmmagent.exe
```

#### Building the 32-bit agent
```
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o out\rmmagent.exe
```

### Signing the agent

See [CODESIGN](CODESIGN.md) for more information.

### Agent installation and deployment

From the server, choose the 'Manual' method when generating an agent. Copy the command line arguments and pass them to the binary. You can also modify the PowerShell script that's available in the dashboard.

#### Creating an installation (setup) file

Packaging the `64-bit` agent with Inno Setup:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup.iss
```

Packaging the `32-bit` agent with Inno Setup:
```
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" build\setup-x86.iss
```

### Updating the agent

Any automatic updates sent by the server will be rejected. A system administrator can still update agents remotely by deploying a custom script/command (instructions to be provided soon, or feel free to use the CLI if you know what you're doing).

### Branding the agent

Take a look at the constants at the top of every file to change the displayed names.
