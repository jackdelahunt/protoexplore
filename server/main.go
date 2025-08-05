package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackdelahunt/protoexplore/explore"
	"google.golang.org/grpc"
	"log"
	"math/rand/v2"
	"net"
)

type ExploreServer struct {
	explore.UnimplementedExploreServiceServer
}

func (s ExploreServer) ListLikedYou(ctx context.Context, request *explore.ListLikedYouRequest) (*explore.ListLikedYouResponse, error) {
	return nil, nil
}
func (s ExploreServer) ListNewLikedYou(ctx context.Context, request *explore.ListLikedYouRequest) (*explore.ListLikedYouResponse, error) {
	return nil, nil
}
func (s ExploreServer) CountLikedYou(ctx context.Context, request *explore.CountLikedYouRequest) (*explore.CountLikedYouResponse, error) {
	return &explore.CountLikedYouResponse{Count: 100}, nil
}
func (s ExploreServer) PutDecision(ctx context.Context, request *explore.PutDecisionRequest) (*explore.PutDecisionResponse, error) {
	return nil, nil
}

func NewServer() *ExploreServer {
	return &ExploreServer{}
}

func GenerateUsers(ctx context.Context, db *pgx.Conn, count int) ([]User, error) {
	users := make([]User, count)

	for i := 0; i < count; i++ {
		user, err := InsertUser(ctx, db)
		if err != nil {
			return nil, err
		}

		users[i] = user
	}

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

	return decisions, nil
}

func main() {
	ctx := context.Background()

	db, err := ConnectToDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close(ctx)

	if true { // generate data
		users, err := GenerateUsers(ctx, db, 200)
		if err != nil {
			log.Fatalf("failed to generate users: %v", err)
		}

		_, err = GenerateDecisions(ctx, db, users)
		if err != nil {
			log.Fatalf("failed to generate decisions: %v", err)
		}
	}

	if false { // make queries
		u := User{
			Id: "f2c1dfad-a3d1-427d-8418-da12024f7466",
		}

		users, err := GetAllLikes(ctx, db, u)
		if err != nil {
			log.Fatalf("failed to get all likes: %v", err)
		}

		fmt.Printf("All likes received %v\n", len(users))
		for _, user := range users {
			fmt.Println(user)
		}

		users, err = GetAllLikesOneWay(ctx, db, u)
		if err != nil {
			log.Fatalf("failed to get all likes: %v", err)
		}

		fmt.Printf("All likes received but not sent %v\n", len(users))
		for _, user := range users {
			fmt.Println(user)
		}

		count, err := GetLikeCount(ctx, db, u)
		if err != nil {
			log.Fatalf("failed to get like counts: %v", err)
		}

		fmt.Printf("Like received count %v\n", count)
	}

	return

	user, err := InsertUser(ctx, db)
	if err != nil {
		log.Fatalf("failed to insert new user: %v", err)
	}

	fmt.Printf("Created user: %v\n", user)

	d := Decision{
		FromUser: "a505623b-0b4d-47d8-ace3-f0235930345c",
		ToUser:   "232679e1-d0a0-4b3f-a882-278c53ee14db",
		Liked:    false,
	}

	nd, err := InsertDecision(ctx, db, d)
	if err != nil {
		log.Fatalf("failed to insert new decision: %v", err)
	}

	fmt.Printf("Created decision: %v\n", nd)

	return

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	println("Listening on " + listener.Addr().String())

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	explore.RegisterExploreServiceServer(grpcServer, NewServer())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpc server encountered an error : %v", err)
	}
}
