package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackdelahunt/protoexplore/explore"
	"log"
)

type ExploreServer struct {
	explore.UnimplementedExploreServiceServer

	Port     string
	Database *pgx.Conn
}

func (s ExploreServer) ListLikedYou(ctx context.Context, request *explore.ListLikedYouRequest) (*explore.ListLikedYouResponse, error) {
	user := User{
		Id: UUID(request.RecipientUserId),
	}

	var users []User
	var paginationToken string
	var err error

	if request.PaginationToken == nil {
		log.Printf("ListLikedYou request: [page=nil] [id=%v]", request.RecipientUserId)
		users, paginationToken, err = GetAllLikesStart(ctx, s.Database, user)
	} else {
		log.Printf("ListLikedYou request: [page=%v] [id=%v]", *request.PaginationToken, request.RecipientUserId)
		users, paginationToken, err = GetAllLikesPaged(ctx, s.Database, user, *request.PaginationToken)
	}

	if err != nil {
		log.Printf("error getting like list: %v", err)
		return nil, err
	}

	likers := make([]*explore.ListLikedYouResponse_Liker, len(users))

	for i, user := range users {
		likers[i] = &explore.ListLikedYouResponse_Liker{
			ActorId:       string(user.Id),
			UnixTimestamp: uint64(user.CreatedAt.Unix()),
		}
	}

	// If we didn't get all of the users that we wanted then it means we can
	// tell the client to not send anymore request by returning nil instead of the token.
	// It would be the case that a wasted request is made if the last user just
	// so happens to be the last in a full page. The next response would then have 0 users
	// but that would still trigger this to send a nil token back and the client
	// shouldn't send anymore requests after that
	var paginationTokenPtr *string

	if len(likers) == PaginationSize {
		paginationTokenPtr = &paginationToken
	}

	response := &explore.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: paginationTokenPtr,
	}

	return response, nil
}
func (s ExploreServer) ListNewLikedYou(ctx context.Context, request *explore.ListLikedYouRequest) (*explore.ListLikedYouResponse, error) {
	log.Printf("ListNewLikedYou request: [id=%v]", request.RecipientUserId)

	user := User{
		Id: UUID(request.RecipientUserId),
	}

	users, err := GetAllLikesOneWay(ctx, s.Database, user)
	if err != nil {
		log.Printf("error getting new like list: %v", err)
		return nil, err
	}

	likers := make([]*explore.ListLikedYouResponse_Liker, len(users))

	for i, user := range users {
		likers[i] = &explore.ListLikedYouResponse_Liker{
			ActorId:       string(user.Id),
			UnixTimestamp: uint64(user.CreatedAt.Unix()),
		}
	}

	response := &explore.ListLikedYouResponse{
		Likers:              likers,
		NextPaginationToken: nil,
	}

	return response, nil
}
func (s ExploreServer) CountLikedYou(ctx context.Context, request *explore.CountLikedYouRequest) (*explore.CountLikedYouResponse, error) {
	log.Printf("CountLikedYou request: [id=%v]", request.RecipientUserId)

	user := User{
		Id: UUID(request.RecipientUserId),
	}

	count, err := GetLikeCount(ctx, s.Database, user)
	if err != nil {
		log.Printf("error getting liked count: %v", err)
		return nil, err
	}

	response := &explore.CountLikedYouResponse{Count: count}

	return response, nil
}
func (s ExploreServer) PutDecision(ctx context.Context, request *explore.PutDecisionRequest) (*explore.PutDecisionResponse, error) {
	log.Printf("PutDecision request: [from=%v] [to=%v]", request.ActorUserId, request.RecipientUserId)

	decisionRequest := Decision{
		FromUser: UUID(request.ActorUserId),
		ToUser:   UUID(request.RecipientUserId),
		Liked:    request.LikedRecipient,
	}

	_, err := InsertDecision(ctx, s.Database, decisionRequest)
	if err != nil {
		return nil, err
	}

	mutualLike := false

	// if this was a like then check for the same decision but from the to_user
	// to the from_user in the original decision, if there was no like then dont bother
	if request.LikedRecipient {
		oppositeDecision := Decision{
			FromUser: UUID(request.RecipientUserId),
			ToUser:   UUID(request.ActorUserId),
		}

		exists, err := GetDecision(ctx, s.Database, oppositeDecision, &oppositeDecision)
		if err != nil {
			return nil, err
		}

		if exists && oppositeDecision.Liked {
			mutualLike = true
		}
	}

	response := &explore.PutDecisionResponse{MutualLikes: mutualLike}

	return response, nil
}
