# Auth API

```
gin --appPort 8280 --port 8280 --all --immediate --path ./src or air
```

| pk   | sk   |       |           |      |           |
| ---- | ---- | ----- | --------- | ---- | --------- |
| user | user | email | password  | data | createdAt |
| org  | org  | name  | createdBy |      |           |
| user | org  | role  |           |      |           |
| role | role | name  |           |      |           |
