package main

import (
	"flag"
	"log"

	"yunion.io/x/ocadm/pkg/util/mysql"
)

var (
	host     = flag.String("host", "", "sql host")
	port     = flag.Int("port", 3306, "sql port")
	user     = flag.String("user", "root", "sql user")
	password = flag.String("password", "", "sql password")
	db       = flag.String("db", "db", "sql test db")
)

func main() {
	flag.Parse()
	conn, err := mysql.NewConnection(*host, *port, *user, *password)
	if err != nil {
		panic(err)
	}
	if err := conn.CheckHealth(); err != nil {
		log.Fatalf("Not health: %v", err)
	}
	log.Printf("Health ok.")
	exists, err := conn.IsDatabaseExists(*db)
	if err != nil {
		log.Fatalf("Check exists error: %v", err)
	}
	log.Printf("db %s exists: %v", *db, exists)

	testDB := "gotestdb"
	if err := conn.CreateDatabase(testDB); err != nil {
		log.Fatalf("Create database %v", err)
	}
	testUser := "lzxlzx"
	testPass := "lzxlzx"
	if err := conn.CreateUser(testUser, testPass, testDB); err != nil {
		log.Fatalf("Create user: %v", err)
	}
	if err := conn.DropDatabase(testDB); err != nil {
		log.Fatalf("Drop database: %v", err)
	}
	ret, err := conn.IsGrantPrivUser(*user, "")
	log.Printf("%v, error: %v", ret, err)
	ret, err = conn.IsGrantPrivUser(testUser, "")
	log.Printf("%v, error: %v", ret, err)
	conn.DropUser(testUser)
}
