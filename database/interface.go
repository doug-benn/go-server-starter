package database

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-benn/go-server-starter/models"
)

// import (
// 	"fmt"

// 	"github.com/doug-benn/go-server-starter/models"
// )

func (db *postgresDatabase) GetAllComments() ([]models.Comment, error) {
	comments := []models.Comment{}
	fmt.Println("Gettings All Data")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rows, err := db.Sql.QueryContext(ctx, `SELECT * FROM comments;`)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		comment := models.Comment{}

		if err := rows.Scan(
			&comment.ID,
			&comment.Comment,
		); err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println(comment)

		comments = append(comments, comment)
	}

	fmt.Printf("Data has value %+v\n\n", comments)
	return comments, nil

}
