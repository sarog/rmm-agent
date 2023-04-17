#define MyAppName "RMM Agent"
#define MyAppVersion "1.7.2"
#define MyAppPublisher "ACME"
#define MyAppURL "https://example.com"
#define MyAppExeName "rmmagent.exe"
#define NSSM "nssm-x86.exe"
#define MESHEXE "meshagent-x86.exe"
#define SALTUNINSTALL "{sd}\salt\uninst.exe"
#define SALTDIR "{sd}\salt"
#define MESHDIR "{sd}\Program Files\Mesh Agent"
#define SERVICE_AGENT_NAME "rmmagent"
#define SERVICE_RPC_NAME "rpcagent"

[Setup]
AppId={{0D34D278-5FAF-4159-A4A0-4E2D2C08139D}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName="{sd}\Program Files\RMMAgent"
DisableDirPage=yes
SetupLogging=yes
DisableProgramGroupPage=yes
OutputBaseFilename=winagent-v{#MyAppVersion}-x86
SetupIconFile=onit.ico
WizardSmallImageFile=onit.bmp
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma
SolidCompression=yes
WizardStyle=modern
RestartApplications=no
CloseApplications=no
MinVersion=6.0

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "..\out\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion;
Source: "nssm-x86.exe"; DestDir: "{app}"

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent runascurrentuser

[UninstallRun]
Filename: "{app}\{#NSSM}"; Parameters: "stop {#SERVICE_AGENT_NAME}"; RunOnceId: "stoprmmagent";
Filename: "{app}\{#NSSM}"; Parameters: "remove {#SERVICE_AGENT_NAME} confirm"; RunOnceId: "removermmagent";
Filename: "{app}\{#NSSM}"; Parameters: "stop {#SERVICE_RPC_NAME}"; RunOnceId: "stoprmmrpc";
Filename: "{app}\{#NSSM}"; Parameters: "remove {#SERVICE_RPC_NAME} confirm"; RunOnceId: "removermmrpc";
Filename: "{app}\{#MyAppExeName}"; Parameters: "-m cleanup"; RunOnceId: "cleanuprm";
Filename: "{cmd}"; Parameters: "/c taskkill /F /IM {#MyAppExeName}"; RunOnceId: "killrmmagent";
Filename: "{#SALTUNINSTALL}"; Parameters: "/S"; RunOnceId: "saltrm"; Check: FileExists(ExpandConstant('{sd}\salt\uninst.exe'));
Filename: "{app}\{#MESHEXE}"; Parameters: "-fulluninstall"; RunOnceId: "meshrm";

[UninstallDelete]
Type: filesandordirs; Name: "{app}";
Type: filesandordirs; Name: "{#SALTDIR}"; Check: DirExists(ExpandConstant('{sd}\salt'));
Type: filesandordirs; Name: "{#MESHDIR}";

[Code]
function InitializeSetup(): boolean;
var
  ResultCode: Integer;
begin
  Exec('cmd.exe', '/c net stop {#SERVICE_AGENT_NAME}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Stopping RMM agent service: ' + IntToStr(ResultCode));
  Exec('cmd.exe', '/c net stop checkrunner', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec('cmd.exe', '/c net stop {#SERVICE_RPC_NAME}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Stopping agent RPC service: ' + IntToStr(ResultCode));
  Exec('cmd.exe', '/c taskkill /F /IM {#MyAppExeName}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('taskkill: ' + IntToStr(ResultCode));

  Result := True;
end;

procedure DeinitializeSetup();
var
  ResultCode: Integer;
begin
  Exec('cmd.exe', '/c net start {#SERVICE_AGENT_NAME} && ping 127.0.0.1 -n 2', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Starting RMM agent service: ' + IntToStr(ResultCode));
  Exec('cmd.exe', '/c net start {#SERVICE_RPC_NAME}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Starting agent RPC service: ' + IntToStr(ResultCode));
end;

