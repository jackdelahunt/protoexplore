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

var GLobalClientID string

func ReadClientId() (string, error) {
	path := "bin/client.id"

	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(bytes), " \n\r\t"), nil
}

func NewClient(host string, port string) (explore.ExploreServiceClient, *grpc.ClientConn, error) {
	url := fmt.Sprintf("%s:%s", host, port)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	connection, err := grpc.NewClient(url, opts...)
	if err != nil {
		log.Fatalf("failed to create connection to %v: %v", url, err)
	}

	client := explore.NewExploreServiceClient(connection)

	log.Printf("created service client connected to %v", url)

	return client, connection, nil
}

func EveryLikeListMenuOption(ctx context.Context, client explore.ExploreServiceClient, scanner *bufio.Scanner) error {
	var paginationToken *string = nil

	for {
		request := explore.ListLikedYouRequest{RecipientUserId: GLobalClientID, PaginationToken: paginationToken}

		response, err := client.ListLikedYou(ctx, &request)
		if err != nil {
			return err
		}

		for _, liker := range response.GetLikers() {
			fmt.Printf("%v liked you at\n", liker.ActorId)
		}

		paginationToken = response.NextPaginationToken
		if paginationToken == nil {
			fmt.Println("That was all of your likes")
			break
		}

		_ = StringInputWithPrompt(scanner, "press enter for next page")
	}

	return nil
}

func NewLikeListMenuOption(ctx context.Context, client explore.ExploreServiceClient) error {
	request := explore.ListLikedYouRequest{RecipientUserId: GLobalClientID, PaginationToken: nil}

	response, err := client.ListNewLikedYou(ctx, &request)
	if err != nil {
		return err
	}

	fmt.Printf("%v people liked you without a like back\n", len(response.GetLikers()))

	for _, liker := range response.GetLikers() {
		fmt.Printf("%v liked you\n", liker.ActorId)
	}

	return nil
}

func CountAllLikesMenuOption(ctx context.Context, client explore.ExploreServiceClient) error {
	request := explore.CountLikedYouRequest{RecipientUserId: GLobalClientID}

	response, err := client.CountLikedYou(ctx, &request)
	if err != nil {
		return err
	}

	fmt.Printf("You received %v likes!\n", response.Count)

	return nil
}

func MatchFromLikeListMenuOption(ctx context.Context, client explore.ExploreServiceClient, scanner *bufio.Scanner) error {
	fmt.Printf("Use y/n/q to match, pass or stop matching. Some may have already passed but give them a second chance\n")

	request := explore.ListLikedYouRequest{RecipientUserId: GLobalClientID, PaginationToken: nil}

	listResponse, err := client.ListNewLikedYou(ctx, &request)
	if err != nil {
		return err
	}

	i := 0 // manually handle i because bad input
	likers := listResponse.GetLikers()

loop:
	for i < len(likers) {
		liker := likers[i]

		fmt.Printf("do you like %v?\n", liker.ActorId)
		input := StringInputWithPrompt(scanner, "y/n/q > ")

		if len(input) != 1 {
			fmt.Println("Just one character please!")
			continue
		}

		liked := false

		switch input[0] {
		case 'y':
			liked = true
			fallthrough
		case 'n':
			i += 1

			decisionRequest := explore.PutDecisionRequest{
				ActorUserId:     GLobalClientID,
				RecipientUserId: liker.ActorId,
				LikedRecipient:  liked,
			}

			decisionResponse, err := client.PutDecision(ctx, &decisionRequest)
			if err != nil {
				return err
			}

			// this should always be true but at least it not being true is supported
			if decisionResponse.MutualLikes {
				fmt.Println("It's a match congrats!!")
			}
		case 'q':
			fmt.Println("Back to main menu")
			break loop
		default:
			fmt.Println("Incorrect input! only \"y\" \"n\" and \"q\" allowed")
			continue
		}
	}

	return nil
}

func StringInputWithPrompt(scanner *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func IntInputWithPrompt(scanner *bufio.Scanner, prompt string) (int, bool) {
	input := StringInputWithPrompt(scanner, prompt)
	n, err := strconv.Atoi(input)

	return n, err == nil
}

func main() {
	log.SetPrefix("[Client]: ")

	/* Read Client ID file */
	id, err := ReadClientId()
	if err != nil {
		log.Fatalf("failed to read client id %v", err)
	}

	log.Printf("client id file found, now using [id=%v]", id)

	// this is stored globally because unlike the server the client implementation
	// is done for us so having it global is an easy way to access it from the client
	GLobalClientID = id

	client, connection, err := NewClient(os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Fatalf("failed to create client instance: %v", err)
	}

	defer connection.Close()

	/* Run interactive CLI in a REPL  */
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n--- Protoexplore ---")
		fmt.Println("1) See everyone who liked you")
		fmt.Println("2) See likes from people you haven't yet")
		fmt.Println("3) Get your total likes")
		fmt.Println("4) Match with people who liked you")
		fmt.Println("5) Exit")

		choice, ok := IntInputWithPrompt(scanner, "> ")
		if !ok {
			continue
		}

		switch choice {
		case 1:
			err := EveryLikeListMenuOption(context.Background(), client, scanner)
			if err != nil {
				log.Printf("uh oh! we may have broke something, try again later: %v", err)
			}
		case 2:
			err := NewLikeListMenuOption(context.Background(), client)
			if err != nil {
				log.Printf("uh oh! we may have broke something, try again later: %v", err)
			}
		case 3:
			err := CountAllLikesMenuOption(context.Background(), client)
			if err != nil {
				log.Printf("uh oh! we may have broke something, try again later: %v", err)
			}
		case 4:
			err := MatchFromLikeListMenuOption(context.Background(), client, scanner)
			if err != nil {
				log.Printf("uh oh! we may have broke something, try again later: %v", err)
			}
		case 5:
			println("Come back soon! ...exiting")
			os.Exit(0)
		default:
			fmt.Println("That option is not available, please choose between 1 and 5")
		}
	}
}
