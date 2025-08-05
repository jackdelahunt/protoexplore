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

//go:embed queries/get_all_likes_received_start.sql
var getAllLikesReceivedStartSQL string

//go:embed queries/get_all_likes_received_paged.sql
var getAllLikesReceivedPagedSQL string

//go:embed queries/get_all_likes_sent.sql
var getAllLikesSentSQL string

//go:embed queries/get_liked_count.sql
var getLikedCountSQL string

//go:embed queries/get_decision.sql
var getDecisionSQL string

const PaginationSize int = 10

func ConnectToDB(ctx context.Context, host string, user string, password string, database string, port string) (*pgx.Conn, error) {
	url := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", user, password, host, port, database)

	// if the server is started at the same time as the database it may not
	// be ready for connections and the server would just fail. This allows for
	// retrying to establish a connection to make sure if it is not connecting
	// it is something actually going wrong

	retryCount := 15
	retryDelay := 1 * time.Second

	for i := range retryCount {
		db, err := pgx.Connect(ctx, url)
		if err != nil {
			log.Printf("[%v] failed to connect to database, will retry: %v\n", i, err.Error())

			time.Sleep(retryDelay)
			continue
		}

		return db, nil
	}

	return nil, errors.New("failed to many connection attempts")
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
		return Decision{}, err
	}

	return newDecision, nil
}

func GetAllLikesStart(ctx context.Context, db *pgx.Conn, user User) ([]User, string, error) {
	rows, err := db.Query(ctx, getAllLikesReceivedStartSQL, user.Id, PaginationSize)
	if err != nil {
		return nil, "", err
	}

	defer rows.Close()

	users, paginationToken, err := RowsToUserList(rows)
	if err != nil {
		return nil, "", err
	}

	return users, paginationToken, nil
}

func GetAllLikesPaged(ctx context.Context, db *pgx.Conn, user User, paginationToken string) ([]User, string, error) {
	rows, err := db.Query(ctx, getAllLikesReceivedPagedSQL, user.Id, paginationToken, PaginationSize)
	if err != nil {
		return nil, "", err
	}

	defer rows.Close()

	users, newPaginationToken, err := RowsToUserList(rows)
	if err != nil {
		return nil, "", err
	}

	return users, newPaginationToken, nil
}

func GetAllLikesOneWay(ctx context.Context, db *pgx.Conn, user User) ([]User, error) {
	// I couldn't figure out how to do this query how (I think) it is intended
	// where its all done in the query I could only get as far as this where I
	// check for the like both to and from the user and filter it on the server :[

	// This also means I wasn't able to do paging for this has it is making
	// two queries in one. Or atleast I couldn't figure out how to

	var likesRecived []User
	var likesSent []User
	var likesRecivedButNotSent []User

	{
		rows, err := db.Query(ctx, getAllLikesReceivedSQL, user.Id)
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		likesRecived, _, err = RowsToUserList(rows)
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

		likesSent, _, err = RowsToUserList(rows)
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

func GetLikeCount(ctx context.Context, db *pgx.Conn, user User) (uint64, error) {
	var count uint64

	err := db.QueryRow(ctx, getLikedCountSQL, user.Id).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetDecision(ctx context.Context, db *pgx.Conn, decision Decision, out *Decision) (bool, error) {
	err := db.
		QueryRow(ctx, getDecisionSQL, decision.FromUser, decision.ToUser).
		Scan(&out.Id, &out.CreatedAt, &out.FromUser, &out.ToUser, &out.Liked)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func RowsToUserList(rows pgx.Rows) ([]User, string, error) {
	var users []User
	var paginationToken string

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.CreatedAt); err != nil {
			return nil, "", err
		}

		users = append(users, u)
		paginationToken = string(u.Id)
	}

	return users, paginationToken, nil
}
