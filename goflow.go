package main

import (
	"fmt"
	"github.com/adambaumeister/goflow/api"
	"github.com/adambaumeister/goflow/config"
)

func main() {

	gc := config.Read("config.yml")

	fmt.Printf("Starting threads...")
	fe := gc.GetFrontends()
	for _, f := range fe {
		go f.Start()
	}
	api.Start()
}