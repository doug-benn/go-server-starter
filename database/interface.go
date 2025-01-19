package database

import (
	"fmt"

	"github.com/doug-benn/go-server-starter/models"
)

type PostgresController struct {
	db *Database
}

func (p *PostgresController) GetAll() (*[]models.Comment, error) {
	query := `SELECT * FROM comments;`

	comments := []models.Comment{}
	fmt.Println("Gettings All Data")

	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()

	rows, err := p.db.Sql.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := &models.Comment{}
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
	return &comments, nil

}
