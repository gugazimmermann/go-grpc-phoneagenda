package main

import (
	"context"
	"fmt"
	"go-grpc-phonebook/phonebookpb/phonebookpb"
	"io"
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

	//createPerson(c)
	//readPerson(c)
	//updatePerson(c)
	//deletePerson(c)
	listPerson(c)
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
	res, err := c.ReadPerson(context.Background(), &phonebookpb.PersonIdRequest{PersonId: personId})
	if err != nil {
		fmt.Printf("Error while reading the person: %v\n", err)
	}
	fmt.Printf("Person: %v\n", res)
}

func updatePerson(c phonebookpb.PhoneBookServiceClient) {
	// CHANGE TO THE ID THAT YOU RECEIVED WHEN CREATE THE PERSON
	// YOU CAN TRY 605812e409be8dac8d59b5af TO SEE code = NotFound
	// AND xxxx TO SEE code = InvalidArgument
	personId := "60594949c5d0fac6fd42fc11"
	fmt.Printf("Update person with ID: %v\n", personId)
	person := &phonebookpb.Person{
		Id:    personId,
		Name:  "José Augusto Zimmermann de Negreiros",
		Email: "jose.augusto@x-team.com",
		Phones: []*phonebookpb.Person_PhoneNumber{
			{
				Number: "+55 47 98870-4247",
				Type:   phonebookpb.Person_WORK,
			},
		},
	}
	res, err := c.UpdatePerson(context.Background(), &phonebookpb.PersonRequest{Person: person})
	if err != nil {
		fmt.Printf("Error while updating the person: %v\n", err)
	}
	fmt.Printf("Person: %v\n", res)
}

func deletePerson(c phonebookpb.PhoneBookServiceClient) {
	// CHANGE TO THE ID THAT YOU RECEIVED WHEN CREATE THE PERSON
	// YOU CAN TRY 605812e409be8dac8d59b5af TO SEE code = NotFound
	// AND xxxx TO SEE code = InvalidArgument
	personId := "60594949c5d0fac6fd42fc11"
	fmt.Printf("Deleting person with ID: %v\n", personId)
	res, err := c.DeletePerson(context.Background(), &phonebookpb.PersonIdRequest{PersonId: personId})
	if err != nil {
		fmt.Printf("Error while deleting the person: %v\n", err)
	}
	fmt.Printf("Person: %v\n", res)
}

func listPerson(c phonebookpb.PhoneBookServiceClient) {
	fmt.Println("listPerson...")
	stream, err := c.ListPerson(context.Background(), &phonebookpb.ListPersonResquest{})
	if err != nil {
		fmt.Printf("Error while calling ListPerson RPC: %v\n", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened while receive stream: %v\n", err)
		}
		fmt.Println(res.GetPerson())
	}
}
