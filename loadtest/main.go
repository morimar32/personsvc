package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	person "personsvc/generated"

	"google.golang.org/grpc"
)

const (
	address = "0.0.0.0:9090"
)

func main() {

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(8)
	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	go func() {
		loadtest()
		defer wg.Done()
	}()

	wg.Wait()
	fmt.Printf("done. elapsed: %d\n", time.Since(start)/1E9)
}

func loadtest() {
	fnames := []string{"Bob", "John", "Paul", "Matt", "Tim", "Jeff", "Josh", "Scott", "Nick", "Chad", "Alex", "Jessica", "Megan", "Sharon", "James", "Harry"}
	lnames := []string{"Anderson", "Jones", "Smith", "Harris", "Wilson", "Cruise", "Johnson", "Ericson", "Corgan", "Callan", "Martin", "Harrison", "Stone", "Potter", "Lee"}
	total := 0

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := person.NewPersonClient(conn)

	for _, lname := range lnames {
		for _, fname := range fnames {
			//fmt.Printf("%s %s\n", fname, lname)
			req := &person.AddPersonRequest{
				FirstName: fname,
				LastName:  lname,
			}
			rec, err := c.AddPerson(context.Background(), req)
			if err != nil {
				log.Fatal(err)
			}
			pr := &person.PersonRequest{
				Id: rec.Id,
			}
			_, err = c.GetPerson(context.Background(), pr)
			if err != nil {
				log.Fatal(err)
			}

			_, err = c.DeletePerson(context.Background(), pr)
			if err != nil {
				log.Fatal(err)
			}
			total++
		}
	}
	fmt.Printf("run done: %d\n", total)
}
