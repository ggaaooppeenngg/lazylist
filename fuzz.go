// +build gofuzz

package lazylist

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

func Fuzz(data []byte) int {
	rand.Seed(time.Now().Unix())
	var l = New(func(v1, v2 interface{}) bool {
		i1 := v1.(int)
		i2 := v2.(int)
		return i1 < i2
	})
	var wait sync.WaitGroup
	set := make(map[int]bool)
	wait.Add(len(data))
	for _, d := range data {
		set[int(d)] = true
		if rand.Int()%2 == 0 {
			go func(i int) {
				defer wait.Done()
				l.Add(i)
			}(int(d))
		} else {
			go func(i int) {
				defer wait.Done()
				l.Remove(i)
			}(int(d))
		}
	}
	wait.Wait()
	ints := make([]int, 0, len(set))
	for i := range set {
		ints = append(ints, i)
	}
	is := sort.IntSlice(ints)
	is.Sort()
	iter := l.Iterator()
	for index, integer := range is {
		v, ok := iter.Next()
		if !ok {
			fmt.Println(is)
			panic("length not match")
		}
		if v.(int) != integer {
			fmt.Println(is)
			fmt.Print(index, v, " ", integer)
			panic("value not equal")
		}
	}
	return 1
}
