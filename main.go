package main

import (
	"log"
	"sync"
)

func main() {
	subpub := NewSubPub()

	var wg sync.WaitGroup
	wg.Add(1)
	go StartGRPCServer(subpub, &wg)

	wg.Wait()
	log.Println("Server exited")
}
