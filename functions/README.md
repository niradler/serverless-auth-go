# API

```
$env:GOOS = 'linux'
go build -o build/public/main src/public.go
go build -o build/admin/main src/admin.go
go build -o build/private/main src/private.go
go build -o build/authrizer/main src/authrizer.go

$env:GOOS = 'windows'
go run src/main.go
```
