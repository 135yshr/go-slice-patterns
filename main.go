package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	ID    uint
	Name  string
	Age   uint
	Email string
	City  string
}

// パターンA: 値スライス
type HogeA struct {
	Users []User `json:"users,omitempty"`
}

// パターンB: スライスそのもののポインタ
type FugaB struct {
	Users *[]User `json:"users,omitempty"`
}

// パターンC: 要素ポインタのスライス
type PiyoC struct {
	Users []*User `json:"users,omitempty"`
}

func main() {
	fmt.Println("=== 挙動デモ ===")

	// --- A
	fmt.Println("\n[A] 値スライス")
	a := HogeA{}
	printJSON("nil スライス (omitempty適用)", a)
	a.Users = []User{}
	printJSON("空スライス ([])", a)
	a.Users = []User{{ID: 1, Name: "Alice", Age: 20, Email: "a@example.com", City: "Sendai"}}
	printJSON("要素あり", a)

	a2 := HogeA{Users: append([]User(nil), a.Users...)}
	a2.Users[0].Name = "Changed in a2"
	fmt.Printf("a.Users[0].Name=%q, a2.Users[0].Name=%q\n",
		a.Users[0].Name, a2.Users[0].Name)

	// --- B
	fmt.Println("\n[B] スライスポインタ")
	var b FugaB
	printJSON("nilポインタ（キー省略）", b)
	empty := []User{}
	b.Users = &empty
	printJSON("空配列を明示 ([])", b)
	val := []User{{ID: 2, Name: "Bob", Age: 30, Email: "b@example.com", City: "Kanazawa"}}
	b.Users = &val
	printJSON("要素あり", b)
	(*b.Users)[0].Name = "Bob-Updated"
	fmt.Println("B共有性確認:", (*b.Users)[0].Name)

	// --- C
	fmt.Println("\n[C] 要素ポインタのスライス")
	u1 := &User{ID: 3, Name: "Carol", Age: 40, Email: "c@example.com", City: "Tokyo"}
	c := PiyoC{Users: []*User{u1}}
	other := PiyoC{Users: []*User{u1}}
	c.Users[0].Name = "Carol-Shared"
	fmt.Printf("c.Users[0].Name=%q, other.Users[0].Name=%q\n",
		c.Users[0].Name, other.Users[0].Name)

	c.Users = append(c.Users, nil)
	fmt.Printf("Cの2要素目はnil: %v\n", c.Users[1] == nil)
}

func printJSON(label string, v any) {
	b, _ := json.Marshal(v)
	fmt.Printf("%-24s => %s\n", label, b)
}
