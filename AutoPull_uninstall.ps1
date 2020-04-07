# AutoPull Uninstall script Ceated by : Richard Holst

# Remove registry keys
Remove-Item -recurse -Path registry::HKCR\archivehunter -force
# Remove application dir  
Remove-Item -Recurse "C:\Program Files (x86)\autopull\" -force