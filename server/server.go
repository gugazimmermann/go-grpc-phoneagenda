package main

import (
	"context"
	"fmt"
	"go-grpc-phonebook/phonebookpb/phonebookpb"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// make collection global
var collection *mongo.Collection

type personItem struct {
	ID          primitive.ObjectID                `bson:"_id,omitempty"`
	Name        string                            `bson:"name"`
	Email       string                            `bson:"email"`
	Phones      []*phonebookpb.Person_PhoneNumber `bson:"phones"`
	LastUpdated *timestamppb.Timestamp            `bson:"last_updated,omitempty"`
}

type server struct{}

func main() {
	// log if go crash, with the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// get env vars
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDb := os.Getenv("MONGO_DB")

	// create the mongo context
	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect MongoDB
	mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:27017", mongoUsername, mongoPassword)
	fmt.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatalf("Error Starting MongoDB Client: %v", err)
	}

	collection = client.Database(mongoDb).Collection("phoneBook")

	fmt.Println("Starting Listener...")
	l, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	phonebookpb.RegisterPhoneBookServiceServer(s, &server{})

	// Start a GO Routine
	go func() {
		fmt.Println("PhoneBook Server Started...")
		if err := s.Serve(l); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait to exit (Ctrl+C)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block the channel until the signal is received
	<-ch
	fmt.Println("Stopping PhoneBook Server...")
	s.Stop()
	fmt.Println("Closing Listener...")
	l.Close()
	fmt.Println("Closing MongoDB...")
	client.Disconnect(mongoCtx)
	fmt.Println("All done!")
}

func personToPB(data *personItem) *phonebookpb.Person {
	return &phonebookpb.Person{
		Id:          data.ID.Hex(),
		Name:        data.Name,
		Email:       data.Email,
		Phones:      data.Phones,
		LastUpdated: data.LastUpdated,
	}
}

func (*server) CreatePerson(ctx context.Context, req *phonebookpb.PersonRequest) (*phonebookpb.PersonResponse, error) {
	person := req.GetPerson()
	fmt.Printf("CreatePerson called with: %v\n", person)
	data := &personItem{
		Name:        person.GetName(),
		Email:       person.GetEmail(),
		Phones:      person.GetPhones(),
		LastUpdated: timestamppb.Now(),
	}
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf(" Internal Error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Cannot convert to OID")
	}
	data.ID = oid
	return &phonebookpb.PersonResponse{
		Person: personToPB(data),
	}, nil
}

func (*server) ReadPerson(ctx context.Context, req *phonebookpb.PersonIdRequest) (*phonebookpb.PersonResponse, error) {
	personId := req.GetPersonId()
	fmt.Printf("ReadPerson called with: %v\n", personId)
	oid, err := primitive.ObjectIDFromHex(personId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}
	data := &personItem{}
	res := collection.FindOne(context.Background(), primitive.M{"_id": oid})
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot found person with the specified ID: %v", err))
	}
	return &phonebookpb.PersonResponse{
		Person: personToPB(data),
	}, nil
}

func (*server) UpdatePerson(ctx context.Context, req *phonebookpb.PersonRequest) (*phonebookpb.PersonResponse, error) {
	person := req.GetPerson()
	fmt.Printf("CreatePerson called with: %v\n", person)
	oid, err := primitive.ObjectIDFromHex(person.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}
	data := &personItem{
		ID:          oid,
		Name:        person.GetName(),
		Email:       person.GetEmail(),
		Phones:      person.GetPhones(),
		LastUpdated: timestamppb.Now(),
	}
	_, err = collection.ReplaceOne(context.Background(), primitive.M{"_id": oid}, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update person: %v", err))
	}
	return &phonebookpb.PersonResponse{
		Person: personToPB(data),
	}, nil
}

func (*server) DeletePerson(ctx context.Context, req *phonebookpb.PersonIdRequest) (*phonebookpb.DeleteResponse, error) {
	personId := req.GetPersonId()
	fmt.Printf("DeletePerson called with: %v\n", personId)
	oid, err := primitive.ObjectIDFromHex(personId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}
	res, err := collection.DeleteOne(context.Background(), primitive.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete person: %v", err))
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete person: %v", err))
	}
	return &phonebookpb.DeleteResponse{
		Deleted: res.DeletedCount,
	}, nil
}

func (*server) ListPerson(req *phonebookpb.ListPersonResquest, stream phonebookpb.PhoneBookService_ListPersonServer) error {
	fmt.Println("ListPerson start stream")
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown Internal Error: %v", err))
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &personItem{}
		if err := cur.Decode(data); err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Cannot decoding data: %v", err))
		}
		stream.Send(&phonebookpb.PersonResponse{Person: personToPB(data)})
	}
	if err = cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown Internal Error: %v", err))
	}
	return nil
}
