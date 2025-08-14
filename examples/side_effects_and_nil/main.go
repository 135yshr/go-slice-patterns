// examples/side_effects_and_nil/main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

type User struct {
	ID    int
	Name  string
	Email string
	City  string
}

func main() {
	fmt.Println("=== 1) 共有参照による副作用の例 ===")
	sideEffectsDemo()

	fmt.Println("\n=== 2) nil要素の危険（panicやnull混入）の例 ===")
	nilPitfallsDemo()

	fmt.Println("\n=== 3) 安全な書き方（防衛的コピー・nil除去・イミュータブル化） ===")
	safePatternsDemo()
}

// ----------------------------------------
// 1) 共有参照による副作用の例
// ----------------------------------------
func sideEffectsDemo() {
	// 元データ（値）からポインタスライスを作る
	src := []User{
		{ID: 1, Name: "Alice", Email: "a@example.com", City: "Sendai"},
		{ID: 2, Name: "Bob", Email: "b@example.com", City: "Kanazawa"},
	}
	ptrs := toPtrSlice(src)

	// 「フィルタ」などで別のスライスを作るが、要素は同じポインタ参照
	onlySendai := filterPtr(ptrs, func(u *User) bool { return u != nil && u.City == "Sendai" })

	// 片方を更新すると、もう片方にも影響する（共有参照ゆえ）
	fmt.Printf("before: ptrs[0].Name=%q, onlySendai[0].Name=%q\n", ptrs[0].Name, onlySendai[0].Name)
	onlySendai[0].Name = "Alice-Updated"
	fmt.Printf("after : ptrs[0].Name=%q, onlySendai[0].Name=%q  <-- どちらも変わる\n", ptrs[0].Name, onlySendai[0].Name)

	// 「コピー」を作った気になってもポインタの配列だと浅いコピーになりがち
	shallow := append([]*User(nil), ptrs...) // スライス自体は複製だが要素は同じポインタ
	shallow[1].City = "Tokyo"
	fmt.Printf("shallow update -> ptrs[1].City=%q  <-- 共有参照のまま\n", ptrs[1].City)
}

// ----------------------------------------
// 2) nil要素の危険（panicやnull混入）の例
// ----------------------------------------
func nilPitfallsDemo() {
	ptrs := []*User{
		{ID: 1, Name: "Alice", Email: "a@example.com", City: "Sendai"},
		nil, // ← バグや中間処理の結果で混入しがち
		{ID: 3, Name: "Carol", Email: "c@example.com", City: "Nagoya"},
	}

	// 2-1) 単純アクセスでpanic
	// fmt.Println(ptrs[1].Name) // ← runtime error: invalid memory address or nil pointer dereference

	// 2-2) sortでLess関数がnilを想定してないとpanicしうる
	// 安全でない例（コメントアウト）:
	// sort.Slice(ptrs, func(i, j int) bool { return ptrs[i].Name < ptrs[j].Name })

	// nil対応したLess関数（nilは後ろへ）
	sort.Slice(ptrs, func(i, j int) bool {
		ui, uj := ptrs[i], ptrs[j]
		switch {
		case ui == nil && uj == nil:
			return false
		case ui == nil:
			return false
		case uj == nil:
			return true
		default:
			return ui.Name < uj.Name
		}
	})
	fmt.Println("sort ok（nilは末尾）:", ptrNames(ptrs))

	// 2-3) JSONにそのまま流すと null が混ざる
	out, _ := json.Marshal(map[string]any{"users": ptrs})
	fmt.Println("JSON(そのまま):", string(out)) // ...,"users":[{...},null,{...}]

	// nilを除去してからJSONへ
	cleaned := compactNonNil(ptrs)
	out2, _ := json.Marshal(map[string]any{"users": cleaned})
	fmt.Println("JSON(nil除去):", string(out2))
}

// ----------------------------------------
// 3) 安全な書き方（防衛的コピー・nil除去・イミュータブル化）
// ----------------------------------------
func safePatternsDemo() {
	// 3-1) フィルタ結果を「独立」させる（防衛的ディープコピー）
	src := []*User{
		{ID: 1, Name: "Alice", City: "Sendai"},
		{ID: 2, Name: "Bob", City: "Kanazawa"},
		{ID: 3, Name: "Carol", City: "Sendai"},
	}

	// 共有参照にしない版（User値をコピーして新しいポインタを作る）
	copied := filterPtrDeepCopy(src, func(u *User) bool { return u != nil && u.City == "Sendai" })
	// これを更新してもsrc側に影響しない
	copied[0].Name = "Alice-DeepCopied"
	fmt.Printf("deepcopy update -> src[0].Name=%q, copied[0].Name=%q  <-- 独立\n", src[0].Name, copied[0].Name)

	// 3-2) JSON出力時はnil除去 + 値スライス化（`null`混入回避＆API契約を安定化）
	jsonReady := toValueSlice(compactNonNil(src))
	j, _ := json.MarshalIndent(map[string]any{"users": jsonReady}, "", "  ")
	fmt.Println("JSON(値スライス化):\n" + string(j))

	// 3-3) イミュータブル化の一例（setterを持たず、新規値を返す）
	u := User{ID: 10, Name: "Dave", City: "Osaka"}
	u2 := withCity(u, "Tokyo") // 新しい値を返すだけ
	fmt.Printf("immutable-ish: old=%q new=%q\n", u.City, u2.City)

	// 3-4) バッファの再利用や外部公開では必ずディープコピー
	// APIレスポンスのキャッシュを返すとき等に重要
	cache := []*User{{ID: 100, Name: "X"}, {ID: 101, Name: "Y"}}
	safeExternal := deepCopyPtrSlice(cache) // 外部へ渡す前にディープコピーして独立させる
	safeExternal[0].Name = "X-Changed-Outside"
	fmt.Printf("cache[0].Name=%q  <-- 外部更新の副作用を遮断\n", cache[0].Name)
}

// ----------------------------------------
// ユーティリティ
// ----------------------------------------

func toPtrSlice(vs []User) []*User {
	out := make([]*User, len(vs))
	for i := range vs {
		u := vs[i]          // 新しい変数でアドレスが変わらないように
		out[i] = &u         // &vs[i] だとループ変数の罠にならないが、慣習的にこの形が安全
	}
	return out
}

func toValueSlice(ps []*User) []User {
	out := make([]User, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			continue
		}
		out = append(out, *p)
	}
	return out
}

func deepCopyPtrSlice(ps []*User) []*User {
	out := make([]*User, 0, len(ps))
	for _, p := range ps {
		if p == nil {
			out = append(out, nil)
			continue
		}
		cp := *p
		out = append(out, &cp)
	}
	return out
}

func filterPtr(ps []*User, pred func(*User) bool) []*User {
	out := make([]*User, 0, len(ps))
	for _, p := range ps {
		if pred(p) {
			out = append(out, p) // ← そのまま参照を渡す（副作用が伝播）
		}
	}
	return out
}

func filterPtrDeepCopy(ps []*User, pred func(*User) bool) []*User {
	out := make([]*User, 0, len(ps))
	for _, p := range ps {
		if p == nil || !pred(p) {
			continue
		}
		cp := *p // 値コピーして新規ポインタを作る
		out = append(out, &cp)
	}
	return out
}

func compactNonNil(ps []*User) []*User {
	out := make([]*User, 0, len(ps))
	for _, p := range ps {
		if p != nil {
			out = append(out, p)
		}
	}
	return out
}

func ptrNames(ps []*User) string {
	var b bytes.Buffer
	b.WriteString("[")
	for i, p := range ps {
		if i > 0 {
			b.WriteString(", ")
		}
		if p == nil {
			b.WriteString("nil")
		} else {
			b.WriteString(p.Name)
		}
	}
	b.WriteString("]")
	return b.String()
}

// 値オブジェクト風の「更新」：新しい値を返す
func withCity(u User, city string) User {
	u.City = city
	return u
}
