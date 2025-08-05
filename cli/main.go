package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/jackdelahunt/protoexplore/explore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connection, err := grpc.NewClient("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("failed to create connection: %v", err)
	}

	defer connection.Close()

	client := explore.NewExploreServiceClient(connection)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n--- Main Menu ---")
		fmt.Println("1) See everyone who liked you")
		fmt.Println("2) See likes from people you haven't yet")
		fmt.Println("3) Get your total likes")
		fmt.Println("4) Make a decision")
		fmt.Println("5) Exit")

		fmt.Print("Enter choice: ")
		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		choice, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		switch choice {
		case 3:
			err := countAllLikes(context.Background(), client)
			if err != nil {
				fmt.Printf("Error when counting likes, try again later: %v", err)
			}
		case 5:
			println("Come back soon! ...exiting")
			os.Exit(0)
		default:
			fmt.Println("That option is not available, please choose between 1 and 5")
		}
	}
}

func countAllLikes(ctx context.Context, client explore.ExploreServiceClient) error {
	resp, err := client.CountLikedYou(ctx, &explore.CountLikedYouRequest{})
	if err != nil {
		return err
	}

	fmt.Printf("You received %v likes!\n", resp.Count)

	return nil
}
