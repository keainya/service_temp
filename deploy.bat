go mod tidy
set GOOS=linux
go build -o app.out
set GOOS=windows
ssh root@takemeto.icu "mkdir /apps 2>/dev/null"
ssh root@takemeto.icu "mkdir /apps/service_temp 2>/dev/null"
ssh root@takemeto.icu "rm -f /apps/service_temp/app.out 2>/dev/null"
scp .\app.out root@takemeto.icu:/apps/service_temp/app.out
ssh root@takemeto.icu "chmod +x /apps/service_temp/app.out"
ssh root@takemeto.icu "kill $(ps aux | grep /apps/service_temp/app.out | grep -v grep | awk '{print $2}') 2>/dev/null"
ssh root@takemeto.icu "cd /apps/service_temp && nohup /apps/service_temp/app.out > /dev/null 2>&1 &"