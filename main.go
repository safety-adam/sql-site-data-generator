package main

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	se := &siteEngine{}

	se.host = "is-1824-task-postgres-db-0422-110322.cgxxcldfpfww.us-east-1.rds.amazonaws.com"
	se.port = 5432
	se.user = "task_engine_user"
	se.password = ""
	se.dbname = "task_engine"

	se.orgId = "141e3af2-149f-4ebc-9439-471b9eee8452"
	se.init()
	defer se.db.Close()

	se.reset()

	size := 20
	countries := se.createSites(size, "Country", uuid.Nil)
	for _, country := range countries {
		states := se.createSites(size, "State", country)
		for _, state := range states {
			cities := se.createSites(size, "City", state)
			for _, city := range cities {
				se.createSites(size, "Suburb", city)
				//for _, suburb := range suburbs {
				//	se.createSites(size, "Street", suburb)
				//}
			}
		}
	}
}

type siteEngine struct {
	host     string
	port     int
	user     string
	password string
	dbname   string

	orgId string

	conn string
	db   *sql.DB
}

func (se *siteEngine) init() {
	// connection string
	se.conn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", se.host, se.port, se.user, se.password, se.dbname)

	// open database
	var err error
	se.db, err = sql.Open("postgres", se.conn)
	CheckError(err)
}

func (se *siteEngine) createSites(count int, name string, parent uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID

	var p interface{}
	if parent == uuid.Nil {
		p = "null"
	} else {
		p = fmt.Sprintf("'%s'", parent)
	}

	fmt.Printf("Creating %d %s with parent %s\n", count, name, parent)

	s := "INSERT INTO site_folders (id, created_at, org_id, folder_name, parent_id) VALUES\n"
	for c := 1; c <= count; c++ {
		n := fmt.Sprintf("%s %d", name, c)
		s = s + fmt.Sprintf("(uuid_generate_v4(), current_timestamp, '%s', '%s', %s)", se.orgId, n, p)
		if c < count {
			s = s + ","
		}
		s = s + "\n"
	}
	s = s + "RETURNING id"

	r, e := se.db.Query(s)
	CheckError(e)

	for r.Next() {
		var idResult uuid.UUID
		r.Scan(&idResult)
		ids = append(ids, idResult)
	}

	return ids
}

func (s *siteEngine) reset() {
	fmt.Println("Clearing all sites for org")
	delStatement := `DELETE FROM site_folders WHERE org_id=$1`
	_, e := s.db.Exec(delStatement, s.orgId)
	CheckError(e)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
