Write-Host "Building frontend..."
Set-Location frontend
npm install
npm run build
Set-Location ..

Write-Host "Copying frontend dist into backend..."
Remove-Item -Recurse -Force backend\web\dist -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force backend\web | Out-Null
Copy-Item -Recurse -Force frontend\dist backend\web\dist

Write-Host "Building openbridge.exe..."
Set-Location backend
go mod tidy
go build -o ..\openbridge.exe .
Set-Location ..

Write-Host "Done: openbridge.exe"