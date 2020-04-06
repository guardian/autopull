# make dir  
mkdir "C:\Program Files (x86)\autopull\"
# Download Software
Invoke-WebRequest -Uri "http://andydeliverablesbucket.s3.eu-west-2.amazonaws.com/autopull-win.zip" -OutFile "C:\Program Files (x86)\autopull\autopull-win.zip"
# Unzip
Expand-Archive -Path "C:\Program Files (x86)\autopull\autopull-win.zip" -DestinationPath "C:\Program Files (x86)\autopull"
# set Execution policy
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser -Force