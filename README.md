# Go Slice vs Pointer Patterns

## 含まれるパターン

- A: []User
- B: *[]User
- C: []*User

## 挙動デモ

```bash
go run .
```

## ベンチマーク

```bash
go test -bench . -benchmem
```

## ベンチ内容

- 基本操作: 走査 / コピー / 更新
- JSON: Marshal / JSON Lines
- 実ワークロード例: DTO変換 / フィルタ / ソート / グルーピング
