$platforms = @( @{GOOS="windows"; GOARCH="amd64"; OUT="./binary/monitor-agent-windows.exe"}, @{GOOS="linux"; GOARCH="amd64"; OUT="./binary/monitor-agent-linux"}, @{GOOS="darwin"; GOARCH="amd64"; OUT="./binary/monitor-agent-macos-intel"}, @{GOOS="darwin"; GOARCH="arm64"; OUT="./binary/monitor-agent-macos-arm64"} )
foreach ($p in $platforms) {
    $env:GOOS=$p.GOOS
    $env:GOARCH=$p.GOARCH
    go build -o $p.OUT
}

Remove-Item Env:GOOS
Remove-Item Env:GOARCH
