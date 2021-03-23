package main

import (
	"context"
	"fmt"
	"go-grpc-phonebook/phonebookpb/phonebookpb"
	"log"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Starting Client...")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()
	c := phonebookpb.NewPhoneBookServiceClient(cc)

	// createPerson(c)
	readPerson(c)
}

func createPerson(c phonebookpb.PhoneBookServiceClient) {
	fmt.Println("Creating the person...")
	person := &phonebookpb.Person{
		Name:  "Guga Zimmermann",
		Email: "gugazimmermann@gmail.com",
		Phones: []*phonebookpb.Person_PhoneNumber{
			{
				Number: "+55 47 98870-4247",
				Type:   phonebookpb.Person_MOBILE,
			},
			{
				Number: "+55 47 XXXXX-XXXX",
				Type:   phonebookpb.Person_HOME,
			},
		},
	}
	res, err := c.CreatePerson(context.Background(), &phonebookpb.PersonRequest{Person: person})
	if err != nil {
		fmt.Printf("Error while creating the person: %v\n", err)
	}
	fmt.Printf("Person Created: %v\n", res)
}

func readPerson(c phonebookpb.PhoneBookServiceClient) {
	// CHANGE TO THE ID THAT YOU RECEIVED WHEN CREATE THE PERSON
	// YOU CAN TRY 605812e409be8dac8d59b5af TO SEE code = NotFound
	// AND xxxx TO SEE code = InvalidArgument
	personId := "60594949c5d0fac6fd42fc11"
	fmt.Printf("Reading person with ID: %v\n", personId)
	res, err := c.ReadPerson(context.Background(), &phonebookpb.ReadPersonRequest{PersonId: personId})
	if err != nil {
		fmt.Printf("Error while reading the person: %v\n", err)
	}
	fmt.Printf("Person: %v\n", res)
}
