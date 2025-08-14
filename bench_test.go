package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type DTO struct {
	Identifier string
	AgeGroup   string
}

// ---- データ生成
func genUsers(n int) []User {
	us := make([]User, n)
	for i := 0; i < n; i++ {
		us[i] = User{
			ID:    uint(i + 1),
			Name:  "User_" + strconv.Itoa(i),
			Age:   uint(18 + (i % 50)),
			Email: "user" + strconv.Itoa(i) + "@example.com",
			City:  "City" + strconv.Itoa(i%10),
		}
	}
	return us
}

func genPtrUsers(n int) []*User {
	us := make([]*User, n)
	for i := 0; i < n; i++ {
		u := User{
			ID:    uint(i + 1),
			Name:  "User_" + strconv.Itoa(i),
			Age:   uint(18 + (i % 50)),
			Email: "user" + strconv.Itoa(i) + "@example.com",
			City:  "City" + strconv.Itoa(i%10),
		}
		us[i] = &u
	}
	return us
}

var (
	SinkInt   int
	SinkBytes []byte
	SinkUsers []User
	SinkUPtrs []*User
	SinkDTOs  []DTO
)

// 走査
func BenchmarkIterate_ValueSlice(b *testing.B) {
	src := genUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, u := range src {
			sum += int(u.ID)
		}
		SinkInt = sum
	}
}
func BenchmarkIterate_PtrSlice(b *testing.B) {
	src := genPtrUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for _, u := range src {
			sum += int(u.ID)
		}
		SinkInt = sum
	}
}

// コピー
func BenchmarkCopy_ValueSlice(b *testing.B) {
	src := genUsers(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := append([]User(nil), src...)
		SinkUsers = dst
	}
}
func BenchmarkCopy_PtrSlice(b *testing.B) {
	src := genPtrUsers(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := append([]*User(nil), src...)
		SinkUPtrs = dst
	}
}

// 更新
func BenchmarkUpdate_ValueSlice(b *testing.B) {
	src := genUsers(100000)
	dst := append([]User(nil), src...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst[i%len(dst)].Age++
	}
	SinkUsers = dst
}
func BenchmarkUpdate_PtrSlice(b *testing.B) {
	src := genPtrUsers(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src[i%len(src)].Age++
	}
	SinkUPtrs = src
}

// JSON Marshal
func BenchmarkJSON_Marshal_ValueSlice(b *testing.B) {
	src := genUsers(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, _ := json.Marshal(src)
		SinkBytes = out
	}
}
func BenchmarkJSON_Marshal_PtrSlice(b *testing.B) {
	src := genPtrUsers(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, _ := json.Marshal(src)
		SinkBytes = out
	}
}

// 実ワークロード: DTO変換
func BenchmarkDTOTransform_ValueSlice(b *testing.B) {
	src := genUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dtos := make([]DTO, len(src))
		for j, u := range src {
			dtos[j] = DTO{Identifier: strings.ToLower(u.Email), AgeGroup: groupAge(u.Age)}
		}
		SinkDTOs = dtos
	}
}
func BenchmarkDTOTransform_PtrSlice(b *testing.B) {
	src := genPtrUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dtos := make([]DTO, len(src))
		for j, u := range src {
			dtos[j] = DTO{Identifier: strings.ToLower(u.Email), AgeGroup: groupAge(u.Age)}
		}
		SinkDTOs = dtos
	}
}

// 実ワークロード: フィルタ
func BenchmarkFilter_ValueSlice(b *testing.B) {
	src := genUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var filtered []User
		for _, u := range src {
			if u.City == "City5" {
				filtered = append(filtered, u)
			}
		}
		SinkUsers = filtered
	}
}
func BenchmarkFilter_PtrSlice(b *testing.B) {
	src := genPtrUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var filtered []*User
		for _, u := range src {
			if u.City == "City5" {
				filtered = append(filtered, u)
			}
		}
		SinkUPtrs = filtered
	}
}

// 実ワークロード: ソート
func BenchmarkSort_ValueSlice(b *testing.B) {
	src := genUsers(20000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Slice(src, func(i, j int) bool { return src[i].Email < src[j].Email })
	}
}
func BenchmarkSort_PtrSlice(b *testing.B) {
	src := genPtrUsers(20000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Slice(src, func(i, j int) bool { return src[i].Email < src[j].Email })
	}
}

// 実ワークロード: グルーピング
func BenchmarkGroupByCity_ValueSlice(b *testing.B) {
	src := genUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := make(map[string][]User)
		for _, u := range src {
			group[u.City] = append(group[u.City], u)
		}
		SinkInt = len(group)
	}
}
func BenchmarkGroupByCity_PtrSlice(b *testing.B) {
	src := genPtrUsers(50000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := make(map[string][]*User)
		for _, u := range src {
			group[u.City] = append(group[u.City], u)
		}
		SinkInt = len(group)
	}
}

func groupAge(age uint) string {
	switch {
	case age < 20:
		return "teen"
	case age < 30:
		return "20s"
	case age < 40:
		return "30s"
	default:
		return "40+"
	}
}

// JSON Lines風
func BenchmarkJSONLines_Value(b *testing.B) {
	src := genUsers(20000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		for _, u := range src {
			_ = enc.Encode(u)
		}
		SinkBytes = buf.Bytes()
	}
}
func BenchmarkJSONLines_Ptr(b *testing.B) {
	src := genPtrUsers(20000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		for _, u := range src {
			_ = enc.Encode(u)
		}
		SinkBytes = buf.Bytes()
	}
}
