package db

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv" // package used to read the .env file
	_ "github.com/lib/pq"
)

var schema = `
CREATE TABLE users (
  userid SERIAL PRIMARY KEY,
  name TEXT,
  uuid TEXT
);
`

var schemaCheck = `
SELECT EXISTS (
	SELECT FROM
		pg_tables
	WHERE
		schemaname = 'public' AND
		tablename  = 'users'
	);
`

func BuildDBIfNeeded() error {
	dbConn, err := CreateConnection()
	if err != nil {
		return err
	}
	defer dbConn.Close()

	rows, err := dbConn.Queryx(schemaCheck)
	if err != nil {
		return err
	}

	schemaExists := false
	for rows.Next() {
		if err := rows.Scan(&schemaExists); err != nil {
			return err
		}
	}

	if !schemaExists {
		fmt.Println("Creating Schema...")
		dbConn.MustExec(schema)
	}

	return nil
}

// create connection with postgres db
func CreateConnection() (*sqlx.DB, error) {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	// Open the connection
	db, err := sqlx.Connect("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}

	// check the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("CreateConnection: Successfully connected!")
	// return the connection
	return db, nil
}
