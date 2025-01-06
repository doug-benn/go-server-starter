package database

import (
	"fmt"
	"log"
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

func (s *PostgresInterface) GetAccountByID(id int) (*Comment, error) {
	comments := []Comment{}

	rows, err := s.db.sql.Query("SELECT id, comments FROM comments;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.Comment); err != nil {
			log.Fatal(err)
		}
		fmt.Println(comment)
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(comments)

	return nil, nil
}

func (p *PostgresInterface) GetAll() (*[]Comment, error) {
	query := `SELECT * FROM comments;`

	comments := []Comment{}
	fmt.Println("Gettings All Data")

	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()

	rows, err := p.db.sql.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := &Comment{}
		fmt.Println("Gettings All Data")

		fmt.Println(comment)
		if err := rows.Scan(
			&comment.ID,
			&comment.Comment,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		fmt.Println(comment)
		comments = append(comments, *comment)
	}

	fmt.Printf("Data has value %+v\n", comments)

	// comments := []Comment{}

	// comment := &Comment{ID: "1", Comment: "Hello", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	// comments = append(comments, *comment)

	return &comments, nil

}
