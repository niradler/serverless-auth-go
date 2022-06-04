# Build API

```
$env:GOOS = 'linux'
go build -o build src
$env:GOOS = 'windows'
gin --appPort 8280 --port 8280 --all --immediate --path ./src
```

| pk   | sk   |       |           |      |           |
| ---- | ---- | ----- | --------- | ---- | --------- |
| user | user | email | password  | data | createdAt |
| org  | org  | name  | createdBy |      |           |
| user | org  | role  |           |      |           |
| role | role | name  |           |      |           |
