package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/jpurdie/authapi"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbInsert := `INSERT INTO public.companies VALUES (1, now(), now(), NULL, 'admin_company', true);
	INSERT INTO public.roles VALUES (500, 500, 'OWNER');
	INSERT INTO public.roles VALUES (400, 400, 'SUPERUSER');
	INSERT INTO public.roles VALUES (300, 300, 'ADMIN');
	INSERT INTO public.roles VALUES (200, 200, 'SUPERVISOR');
	INSERT INTO public.roles VALUES (100, 100, 'USER');`
	var psn = os.Getenv("DATABASE_URL")
	queries := strings.Split(dbInsert, ";")

	u, err := pg.ParseURL(psn)
	checkErr(err)
	db := pg.Connect(u)
	_, err = db.Exec("SELECT 1")
	checkErr(err)
	createSchema(db, &authapi.Company{}, &authapi.Role{}, &authapi.User{}, &authapi.CompanyUser{})

	for _, v := range queries[0 : len(queries)-1] {
		_, err := db.Exec(v)
		checkErr(err)
	}

	//sec := secure.New(1, nil)

	//userInsert := `INSERT INTO public.users (id, created_at, updated_at, first_name, last_name, password, email, active, role_id, company_id, location_id) VALUES (1, now(),now(), 'Admin', 'admin', '%s', 'johndoe@mail.com', true, 100, 1, 1);`
	//_, err = db.Exec(fmt.Sprintf(userInsert, sec.Hash("admin")))
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createSchema(db *pg.DB, models ...interface{}) {
	for _, model := range models {
		checkErr(db.CreateTable(model, &orm.CreateTableOptions{
			FKConstraints: true,
		}))
	}
}
