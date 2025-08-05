package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackdelahunt/protoexplore/explore"
	"google.golang.org/grpc"
	"log"
	"math/rand/v2"
	"net"
	"os"
)

func GenerateUsers(ctx context.Context, db *pgx.Conn, count int) ([]User, error) {
	users := make([]User, count)

	for i := 0; i < count; i++ {
		user, err := InsertUser(ctx, db)
		if err != nil {
			return nil, err
		}

		users[i] = user
	}

	log.Printf("generated %v users\n", len(users))

	return users, nil
}

func GenerateDecisions(ctx context.Context, db *pgx.Conn, users []User) ([]Decision, error) {
	// there are three states for a decision for any give "from" user and any given "to user"
	// 1. the from user liked the other to user
	// 2. the from user passed on the to user
	// 3. no decision has been made i.e. it does not exists
	var decisions []Decision

	for i, fromUser := range users {
		for j, toUser := range users {
			if i == j {
				continue
			}

			decisionType := 0 // 0 -> none, 1 -> like, 2 -> pass
			f := rand.Float64()

			if f < 0.6 {
				decisionType = 1
			} else if f < 0.8 {
				decisionType = 2
			}

			if decisionType == 1 || decisionType == 2 {
				liked := decisionType == 1

				d := Decision{
					FromUser: fromUser.Id,
					ToUser:   toUser.Id,
					Liked:    liked,
				}

				newDecision, err := InsertDecision(ctx, db, d)
				if err != nil {
					return nil, err
				}

				decisions = append(decisions, newDecision)
			}
		}
	}

	log.Printf("generated %v decisions\n", len(decisions))

	return decisions, nil
}

func RunServer(server *ExploreServer) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", server.Port))
	if err != nil {
		return err
	}

	log.Printf(fmt.Sprintf("listening on %v\n", listener.Addr().String()))

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	explore.RegisterExploreServiceServer(grpcServer, server)

	err = grpcServer.Serve(listener)
	if err != nil {
		return err
	}

	return nil
}

func WriteClientId(users []User) error {
	// cursed way of giving the client an id to use in requests,
	// because I think that I shouldn't change the API
	if len(users) <= 0 {
		return errors.New("no users provided to create client id file")
	}

	path := "bin/client.id"
	id := users[0].Id

	err := os.WriteFile(path, []byte(id), 0644)
	if err != nil {
		return err
	}

	log.Printf("wrote client id to \"%v\"\n", path)

	return nil
}

func main() {
	ctx := context.Background()
	log.SetPrefix("[SERVER]: ")

	/* connect to postgres */
	db, err := ConnectToDB(ctx,
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Printf("connected to database %v\n", db.PgConn().Conn().RemoteAddr().String())

	defer db.Close(ctx)

	/* Generate users and decisions */
	users, err := GenerateUsers(ctx, db, 50)
	if err != nil {
		log.Fatalf("error while generating users: %v", err)
	}

	_, err = GenerateDecisions(ctx, db, users)
	if err != nil {
		log.Fatalf("error while generating decisions: %v", err)
	}

	/* Create client ID file */
	err = WriteClientId(users)
	if err != nil {
		log.Fatalf("error while writing client id file: %v", err)
	}

	/* Run the grpc server */
	server := ExploreServer{
		Port:     os.Getenv("SERVER_PORT"),
		Database: db,
	}

	err = RunServer(&server)
	if err != nil {
		log.Fatalf("the server encountered an error: %v", err)
	}
}
