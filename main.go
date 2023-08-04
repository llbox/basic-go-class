package main

import (
	"basic-go-class/week1"
	"fmt"
)

func main() {
	s1 := []int{1, 2, 3, 4, 5}
	fmt.Println("s1 len:", len(s1))
	fmt.Println("s1 len:", cap(s1))
	s2, err := week1.DeleteAt[int](s1, 4)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("s2 len:", len(s2))
	fmt.Println("s2 len:", cap(s2))
	fmt.Println(s2)
}
