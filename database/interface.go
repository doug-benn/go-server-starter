package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type dbInterface interface {
	GetAll() ([]*Comment, error)
}

type PostgresInterface struct {
	db *Database
}

func NewPostgresInterface(db *Database) (*PostgresInterface, error) {
	// if !db.running {
	// 	return nil, fmt.Errorf("%s", "database is not running")
	// }
	return &PostgresInterface{db: db}, nil
}

////////

func (s *PostgresInterface) GetAccountByID(id int) (*Comment, error) {
	rows, err := s.db.sql.Query("SELECT comment FROM comments;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			log.Fatal(err)
		}
		fmt.Println(title)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return nil, nil
}

func (s *PostgresInterface) GetAccounts() ([]*Comment, error) {
	rows, err := s.db.sql.Query("select * from comments")
	if err != nil {
		return nil, err
	}

	accounts := []*Comment{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Comment, error) {
	comment := new(Comment)
	err := rows.Scan(
		&comment.ID,
		&comment.Comment,
		&comment.CreatedAt,
		&comment.UpdatedAt)

	return comment, err
}

///////

func (p *PostgresInterface) Get(id string) (*Comment, error) {
	query := `
        SELECT *
        FROM comments
        WHERE id = $1`

	var comment Comment

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := p.db.sql.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.Comment,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (p *PostgresInterface) GetAll() (*[]Comment, error) {
	query := `SELECT * FROM comments`

	comments := []Comment{}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := p.db.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := &Comment{}

		if err := rows.Scan(
			&comment.ID,
			&comment.Comment,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, *comment)
	}

	fmt.Printf("Data has value %+v\n", comments)

	// comments := []Comment{}

	// comment := &Comment{ID: "1", Comment: "Hello", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	// comments = append(comments, *comment)

	return &comments, nil

}
