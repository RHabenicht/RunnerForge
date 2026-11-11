[Setup]
AppId={{A3B5E922-60B4-4F52-A8B1-D402DF791510}
AppName=RunnerForge
AppVersion=0.2.0
AppVerName=RunnerForge 0.2.0 development
AppPublisher=NightLife
DefaultDirName={localappdata}\Programs\RunnerForge
DefaultGroupName=RunnerForge
PrivilegesRequired=lowest
MinVersion=10.0
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\dist
OutputBaseFilename=RunnerForge-Setup-0.2.0
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
UninstallDisplayIcon={app}\RunnerForge.exe
DisableProgramGroupPage=yes

[Files]
Source: "..\go\RunnerForge.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\DEVELOPMENT_STATUS.txt"; DestDir: "{app}"; Flags: ignoreversion

[Tasks]
Name: desktopicon; Description: "Create a desktop shortcut"; Flags: unchecked

[Icons]
Name: "{group}\RunnerForge"; Filename: "{app}\RunnerForge.exe"
Name: "{autodesktop}\RunnerForge"; Filename: "{app}\RunnerForge.exe"; Tasks: desktopicon

[Run]
Filename: "{app}\RunnerForge.exe"; Description: "Open RunnerForge"; Flags: nowait postinstall skipifsilent
