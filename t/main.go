package main

import (
	"fmt"

	"github.com/moisespsena-go/proc"

	"github.com/moisespsena/go-default-logger"
)

func main() {
	log := defaultlogger.NewLogger("main")
	pth := "/opt/google/chrome/chrome"
	b, err := proc.NewBinary(pth)
	if err != nil {
		log.Errorf("Open binary failed: %v", err)
		return
	}

	fmt.Println(b.Kill())
}
