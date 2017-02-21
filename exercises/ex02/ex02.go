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

//var ch chan int //= make(chan int, 1)

func thread_1(ch chan int) {
	for i := 0; i < 1000001; i++ {
		j = <-ch
		j++
		ch <- j
	}
}

func thread_2(ch chan int) {
	for i := 0; i < 1000000; i++ {
		j = <-ch
		j--
		ch <- j
	}
}

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	ch := make(chan int, 1)
	j = 0
	ch <- j
	go thread_1(ch)
	go thread_2(ch)
	time.Sleep(1000 * time.Millisecond)
	Println(j)
}
