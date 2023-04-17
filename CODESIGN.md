### Code Signing the agent

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
