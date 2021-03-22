package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/vxcontrol/vxui/server"
)

// PackageVer is semantic version of vxui
var PackageVer string

// PackageRev is revision of vxui build
var PackageRev string

func init() {
	if os.Getenv("DB_USER") == "" ||
		os.Getenv("DB_PASS") == "" ||
		os.Getenv("DB_HOST") == "" ||
		os.Getenv("DB_PORT") == "" ||
		os.Getenv("DB_NAME") == "" {
		panic("DB params is not defined")
	}
	if (os.Getenv("RECAPTCHA_HTML_KEY") == "") !=
		(os.Getenv("RECAPTCHA_API_KEY") == "") {
		panic("use reCaptcha HTML and API keys in the time")
	}
}

func open(args ...string) (*sql.DB, error) {
	addr := fmt.Sprintf("%s:%s@%s/%s?parseTime=true",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
		fmt.Sprintf("tcp(%s:%s)", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")), os.Getenv("DB_NAME"))
	if len(args) > 0 {
		addr = fmt.Sprintf("%s:%s@%s/%s?parseTime=true",
			args[0], args[1], fmt.Sprintf("tcp(%s:%s)", args[2], args[3]), args[4])
	}

	dbcon, err := sql.Open("mysql", addr)
	if err != nil {
		return nil, err
	}
	dbcon.SetMaxIdleConns(0)
	dbcon.SetMaxOpenConns(256)
	return dbcon, nil
}

// MigrateUp is function that receives path to migration dir and runs up ones
func MigrateUp(path string) error {
	log.Println("SQLMigrateUp: ", path)
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()
	migrations := &migrate.FileMigrationSource{
		Dir: path,
	}
	_, err = migrate.Exec(db, "mysql", migrations, migrate.Up)
	return err
}

// MigrateDown is function that receives path to migration dir and runs down ones
func MigrateDown(path string) (err error) {
	log.Println("SQLMigrateDown: ", path)
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()
	migrations := &migrate.FileMigrationSource{
		Dir: path,
	}
	_, err = migrate.Exec(db, "mysql", migrations, migrate.Down)
	return err
}

// @title VXUI Swagger API
// @version 1.0
// @description Swagger API for VXControl backend product.
// @termsOfService http://swagger.io/terms/

// @contact.url https://vxcontrol.com
// @contact.name Dmitry Nagibin
// @contact.email admin@vxcontrol.com

// @license.name Proprietary License
// @license.url https://github.com/vxcontrol/vxui/src/master/LICENSE

// @query.collection.format multi

// @BasePath /api/v1
func main() {
	var version bool
	flag.BoolVar(&version, "version", false, "Print current version of vxui and exit")
	flag.Parse()

	if version {
		fmt.Printf("vxui version is ")
		if PackageVer != "" {
			fmt.Printf("%s", PackageVer)
		} else {
			fmt.Printf("develop")
		}
		if PackageRev != "" {
			fmt.Printf("-%s\n", PackageRev)
		} else {
			fmt.Printf("\n")
		}

		os.Exit(0)
	}

	if err := MigrateUp("migrations"); err != nil {
		log.Fatalln("failed apply of migrations: ", err)
	}
	server.Run()
}
