package main

import (
	"log"
	"os"
)

func main() {
	log := log.New(os.Stdout, "FINANCIFY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	if err := run(); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run() error {
	return nil
}
