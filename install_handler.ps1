New-Item -Path registry::HKCR\ -Name archivehunter -Value "URL:archivehunter Protocol"
New-ItemProperty -Path registry::HKCR\archivehunter -Name "URL Protocol" -Type string
New-Item -Path registry::HKCR\archivehunter -Name "shell"
New-Item -Path registry::HKCR\archivehunter\shell -Name "open"
New-Item -Path registry::HKCR\archivehunter\shell\open -Name "command" -Value "`"c:\Program Files (x86)\autopull\autopull.exe`" %1"
