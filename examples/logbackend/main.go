package main

import (
	"fmt"
	"log"

	"github.com/WiseGrowth/pigeon/backend"
)

type service struct{}

func (s *service) Approve(content []byte) (valid bool, err error) { return true, nil }
func (s *service) Deliver(content []byte) error {
	fmt.Printf("%s", content)
	return nil
}

func main() {
	log.Println("Serving at :5000")
	log.Fatal(backend.ListenAndServe(":5000", &service{}))
}
