package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type UUID string

//go:embed queries/insert_user.sql
var insertUserSQL string

//go:embed queries/insert_decision.sql
var insertDecisionSQL string

//go:embed queries/get_all_likes_received.sql
var getAllLikesReceivedSQL string

//go:embed queries/get_all_likes_sent.sql
var getAllLikesSentSQL string

//go:embed queries/get_liked_count.sql
var getLikedCountSQL string

func ConnectToDB(ctx context.Context) (*pgx.Conn, error) {
	url := "postgres://postgres:@localhost:5432/exploredb"

	db, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error connecting to database: %v", err))
	}

	return db, nil
}

type User struct {
	Id        UUID
	CreatedAt time.Time
}

type Decision struct {
	Id        int
	CreatedAt time.Time
	FromUser  UUID
	ToUser    UUID
	Liked     bool
}

func InsertUser(ctx context.Context, db *pgx.Conn) (User, error) {
	var user User

	err := db.QueryRow(ctx, insertUserSQL).Scan(&user.Id, &user.CreatedAt)
	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	return user, nil
}

func InsertDecision(ctx context.Context, db *pgx.Conn, decision Decision) (Decision, error) {
	var newDecision Decision

	err := db.
		QueryRow(ctx, insertDecisionSQL, decision.FromUser, decision.ToUser, decision.Liked).
		Scan(&newDecision.Id, &newDecision.CreatedAt, &newDecision.FromUser, &newDecision.ToUser, &newDecision.Liked)
	if err != nil {
		log.Fatalf("failed to insert decision: %v", err)
	}

	return newDecision, nil
}

func UpdateDecision(ctx context.Context, db *pgx.Conn, decision Decision) (Decision, error) {
	return decision, nil
}

func GetAllLikes(ctx context.Context, db *pgx.Conn, user User) ([]User, error) {
	rows, err := db.Query(ctx, getAllLikesReceivedSQL, user.Id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users, err := rowsToUserList(rows)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetAllLikesOneWay(ctx context.Context, db *pgx.Conn, user User) ([]User, error) {
	// ChatGPT could of done this in one query but I could only get as
	// far as this where I check for the like both to and from the user and
	// filter it on the server :[

	var likesRecived []User
	var likesSent []User
	var likesRecivedButNotSent []User

	{
		rows, err := db.Query(ctx, getAllLikesReceivedSQL, user.Id)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		likesRecived, err = rowsToUserList(rows)
		if err != nil {
			return nil, err
		}
	}

	{
		rows, err := db.Query(ctx, getAllLikesSentSQL, user.Id)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		likesSent, err = rowsToUserList(rows)
		if err != nil {
			return nil, err
		}
	}

	for _, receivedLike := range likesRecived {
		found := false
		for _, sentLike := range likesSent {
			if receivedLike.Id == sentLike.Id {
				found = true
				break
			}
		}

		if !found {
			likesRecivedButNotSent = append(likesRecivedButNotSent, receivedLike)
		}
	}

	return likesRecivedButNotSent, nil
}

func GetLikeCount(ctx context.Context, db *pgx.Conn, user User) (int, error) {
	var count int

	err := db.QueryRow(ctx, getLikedCountSQL, user.Id).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func rowsToUserList(rows pgx.Rows) ([]User, error) {
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.CreatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

func rowsToUserSet(rows pgx.Rows) (map[UUID]bool, error) {
	var users map[UUID]bool
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.CreatedAt); err != nil {
			return nil, err
		}

		users[u.Id] = true
	}

	return users, nil
}
