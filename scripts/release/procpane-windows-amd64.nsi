; NSIS installer for procpane tray GUI (unsigned; CI release).
; Invoke: makensis /DAPP_EXE=path\to\procpane-app.exe /DOUT_FILE=path\to\setup.exe scripts/release/procpane-windows-amd64.nsi

!include "MUI2.nsh"

!ifndef APP_EXE
  !error "Define APP_EXE to the built procpane-app.exe path"
!endif
!ifndef OUT_FILE
  !error "Define OUT_FILE to the output setup.exe path"
!endif

Name "procpane"
OutFile "${OUT_FILE}"
InstallDir "$PROGRAMFILES64\procpane"
InstallDirRegKey HKLM "Software\procpane" ""
RequestExecutionLevel admin
Unicode true

!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_RUN "$INSTDIR\procpane-app.exe"
!define MUI_FINISHPAGE_RUN_TEXT "Launch procpane"

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

Section "Install"
  SetOutPath "$INSTDIR"
  File "${APP_EXE}"

  WriteRegStr HKLM "Software\procpane" "" $INSTDIR

  WriteUninstaller "$INSTDIR\Uninstall.exe"

  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\procpane" "DisplayName" "procpane"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\procpane" "UninstallString" "$\"$INSTDIR\Uninstall.exe$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\procpane" "InstallLocation" "$INSTDIR"

  CreateDirectory "$SMPROGRAMS\procpane"
  CreateShortCut "$SMPROGRAMS\procpane\procpane.lnk" "$INSTDIR\procpane-app.exe"
  CreateShortCut "$SMPROGRAMS\procpane\Uninstall procpane.lnk" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\procpane-app.exe"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"

  Delete "$SMPROGRAMS\procpane\procpane.lnk"
  Delete "$SMPROGRAMS\procpane\Uninstall procpane.lnk"
  RMDir "$SMPROGRAMS\procpane"

  DeleteRegKey HKLM "Software\procpane"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\procpane"
SectionEnd
