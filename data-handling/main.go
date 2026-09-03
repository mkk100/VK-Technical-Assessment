package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	productsA, errA := fetchSourceA()
	productsB, errB := fetchSourceB()
	productsC, errC := fetchSourceC()

	for _, e := range []error{errA, errB, errC} {
		if e != nil {
			log.Printf("source error: %v", e)
		}
	}

	var all []Product
	all = append(all, productsA...)
	all = append(all, productsB...)
	all = append(all, productsC...)
	all = dedupe(all)

	out, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

// module is a tree of packages, and each folder is a module, main is a special one.
