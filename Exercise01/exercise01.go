/*
global shared int = 0

main:
  spawn thread_1
  spawn thread_2
  join all threads (wait for them to finish)
  print i

thread_1:
  do 1_000_000 times:
    i++
thread_2:
  do 1_000_000 times:
    i--
*/
package main

import (
	. "fmt"
	"time"
)

var j int

func thread_1() {
	for i := 0; i < 1000000; i++ {
		j++
	}
}

func thread_2() {
	for i := 0; i < 1000000; i++ {
		j--
	}
}

func main() {
	j = 0
	go thread_1()
	go thread_2()
	time.Sleep(1000 * time.Millisecond)
	Println(j)
}
