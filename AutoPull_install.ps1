# AutoPull install script Ceated by : Richard Holst

# make dir  
mkdir "C:\Program Files (x86)\autopull\" -force
# Download Software
Invoke-WebRequest -Uri "https://multimedia-public-downloadables.s3.eu-west-1.amazonaws.com/autopull/master/autopull-win.zip" -OutFile "C:\Program Files (x86)\autopull\autopull-win.zip"
# Unzip
Expand-Archive -Path "C:\Program Files (x86)\autopull\autopull-win.zip" -DestinationPath "C:\Program Files (x86)\autopull" -force
# set Execution policy
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser -Force
# set download location for currently logged on user
(Get-Content "C:\Program Files (x86)\autopull\autopull\autopull.yaml").Replace("download_path:","download_path: $env:USERPROFILE\downloads")|
Set-Content "C:\Program Files (x86)\autopull\autopull\autopull.yaml"

# register the protocol handler
New-Item -Path registry::HKCR\ -Name archivehunter -Value "URL:archivehunter Protocol" -force
New-ItemProperty -Path registry::HKCR\archivehunter -Name "URL Protocol" -Type string -force 
New-Item -Path registry::HKCR\archivehunter -Name "shell" -force
New-Item -Path registry::HKCR\archivehunter\shell -Name "open" -force
New-Item -Path registry::HKCR\archivehunter\shell\open -Name "command" -Value "`"c:\Program Files (x86)\autopull\autopull\autopull.exe`" %1" -force