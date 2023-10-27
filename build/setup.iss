#define MyAppName "JetRMM Agent"
#define MyAppVersion "0.1.0"
#define MyAppPublisher "JetRMM"
#define MyAppURL "https://jetrmm.com"
#define MyAppExeName "rmmagent.exe"
#define SERVICE_AGENT_NAME "jetagent"

[Setup]
AppId={{0D34D278-5FAF-4159-A4A0-4E2D2C08139D}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName="{sd}\Program Files\JetAgent"
DisableDirPage=yes
SetupLogging=yes
DisableProgramGroupPage=yes
OutputBaseFilename=winagent-v{#MyAppVersion}
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
Source: "nssm.exe"; DestDir: "{app}"

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent runascurrentuser

[UninstallRun]
Filename: "{app}\{#NSSM}"; Parameters: "stop {#SERVICE_AGENT_NAME}"; RunOnceId: "stoprmmagent";
Filename: "{app}\{#NSSM}"; Parameters: "remove {#SERVICE_AGENT_NAME} confirm"; RunOnceId: "removermmagent";
Filename: "{app}\{#MyAppExeName}"; Parameters: "-m cleanup"; RunOnceId: "cleanuprm";
Filename: "{cmd}"; Parameters: "/c taskkill /F /IM {#MyAppExeName}"; RunOnceId: "killrmmagent";

[UninstallDelete]
Type: filesandordirs; Name: "{app}";

[Code]
function InitializeSetup(): boolean;
var
  ResultCode: Integer;
begin
  Exec('cmd.exe', '/c net stop {#SERVICE_AGENT_NAME}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Stopping RMM agent service: ' + IntToStr(ResultCode));
  Exec('cmd.exe', '/c net stop checkrunner', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Exec('cmd.exe', '/c taskkill /F /IM {#MyAppExeName}', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('taskkill: ' + IntToStr(ResultCode));

  Result := True;
end;

procedure DeinitializeSetup();
var
  ResultCode: Integer;
begin
  Exec('cmd.exe', '/c net start {#SERVICE_AGENT_NAME} && ping 127.0.0.1 -n 2', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Log('Starting JetRMM agent service: ' + IntToStr(ResultCode));
end;

