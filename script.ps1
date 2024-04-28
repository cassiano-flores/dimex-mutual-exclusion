# script.ps1
Start-Process "powershell" -ArgumentList "-Command go run useDIMEX.go 0 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002"
Start-Process "powershell" -ArgumentList "-Command go run useDIMEX.go 1 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002"
Start-Process "powershell" -ArgumentList "-Command go run useDIMEX.go 2 127.0.0.1:5000  127.0.0.1:6001  127.0.0.1:7002"