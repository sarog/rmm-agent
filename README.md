# Fork of the Tactical RMM Agent

This is a fork of [`wh1te909/rmmagent`](https://github.com/wh1te909/rmmagent) version `1.5.1` with the build scripts imported from `1.4.14` before the source repo [deleted them](https://github.com/wh1te909/rmmagent/commit/3fdb2e8c4833e5310840ca79bf394a53f6dbe990).

This `dev` branch is focusing on making the agent on par with 1.7.2 (or latest) features from upstream. It is considered incomplete and unfit for production use, but feel free to test.

**Please note**: downloadable binaries (executables) will not be provided on this GitHub repository as they will be useless. Users are encouraged to [build](#building-the-windows-agent) and [sign](#signing-the-agent) their own executables to guarantee integrity.

## Project goals
- Re-introduce an open source version of the agent while maintaining compat with the server.
- Backport changes from the upstream project, if possible.
- Allow anyone to use and modify the agent for whatever they see fit.
  - This includes using any other future open source RMM server/backend.
- ~~Make the Python dependency optional.~~ ✅

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
- [Go](https://go.dev/dl/) 1.17+
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
env CGO_ENABLED=0 GOARCH=amd64 go build -ldflags "-s -w" -o out\agent.exe
```

#### Building the 32-bit agent
```
goversioninfo
env CGO_ENABLED=0 GOARCH=386 go build -ldflags "-s -w" -o out\agent.exe
```

### Signing the agent

#### Why sign the agent?
- Windows Defender, or any other antivirus for that matter, does not like an application that is able to query & control the host operating system (such as trojan, a backdoor, or an RMM agent!).
- The antivirus _**especially**_ does not like an application that is not _digitally signed_ with a reputable Code Signing certificate.
- An executable that is digitally signed is considered 'vetted' and generally safe for execution.

#### Why sign the agent yourself?
- Signing the agent yourself means you take responsibility for the executable.
- You **reviewed the source code** & built the agent knowing the code was personally vetted by you or your trusted developer.
- Before you distribute & run this newly compiled executable to machines under your responsibility, you want to guarantee it's not tampered with while in transit or while it remains installed on your managed infrastructure.
- The best way to achieve this guarantee is to sign & seal the executable yourself.
- In other words, the agent will be 'enveloped' and 'marked' with your digital signature to enable integrity.
- If the binary is tampered or the signature is invalidated, warnings can and generally will be triggered by the host operating system or the antivirus.
- A signed agent can be verified by the client, by the sysadmin, or most importantly **by you**.

Requirements:
- A coveted _Code Signing_ ("CS") certificate, either purchased from a third-party Certificate Authority of your choosing, or ideally one from your internal Private Key Infrastructure (PKI). If you don't have a PKI, you can self-sign and distribute the public certificate separately.
  - If you are an MSP, you can either purchase a code signing certificate (headache free) or set up your own Trusted Root Certificate Authority (with restrictions) and distribute your CA certificate to your clients (plenty of headaches).
  - If you are a system administrator, just issue yourself a CS from your Enterprise PKI. Your domain already trusts it.
- Microsoft's key signing tool called [SignTool](https://docs.microsoft.com/en-us/windows/win32/seccrypto/signtool) (part of the Windows 10 SDK) or kSoftware's free [kSign](https://www.ksoftware.net/code-signing-certificates/) if you like GUIs (scroll down to the section titled "Download kSign").

Sign the `agent.exe` and optionally the `winagent-x.y.z.exe` setup file.

The following signing & verification examples are from Microsoft's [SignTool documentation](https://docs.microsoft.com/en-us/windows/win32/seccrypto/using-signtool-to-sign-a-file).

#### Examples

Sign the agent with your certificate using a SHA256 algorithm:
```shell
signtool sign /f MyCert.pfx /p MyPassword /fd SHA256 agent.exe 
```

Sign and timestamp the agent:
```shell
signtool sign /f MyCert.pfx /t http://timestamp.digicert.com /fd SHA256 agent.exe
```

Timestamp a file after it was signed:
```shell
signtool timestamp /t http://timestamp.digicert.com agent.exe
```

If you already have your CS certificate loaded in your Windows keystore, you can abbreviate to the following:
```shell
# Automatically chooses an available CS cert from your system:
signtool sign /a /fd SHA256 agent.exe

# Choose a CS cert based on the subject name "My Certificate" found in your User Certificate Store:
signtool sign /n "My Certificate" /fd SHA256 agent.exe 
```

Signature verification is quite simple:
```shell
signtool verify /pa agent.exe
```

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

### Branding your agent

For the time being, it is ill-advised to change any of the TRMM branding unless one understands the consequences. There are plans to centralize and document such changes in a future release to avoid breaking functionality.

If you plan on using this agent with TacticalRMM, avoid changing any of the following identifiers. Doing so will break things (e.g. the agent won't be able to identify itself to the server).
- The names of the two (2) Windows services `tacticalagent` and `tacticalrpc`. However, the service display name & description **can** be changed.
- The `TacticalAgent` folder name in `Program Files`.
- The agent's binary name `tacticalrmm.exe`. Change the `FileDescription` in `versioninfo.json` which is what will show up in Task Manager.
