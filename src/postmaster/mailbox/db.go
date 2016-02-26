package mailbox

import (
	"conduit/log"
	"fmt"
	"github.com/cznic/ql"
	"github.com/kardianos/osext"
	"os"
	"path/filepath"
	"strconv"
)

var DB *ql.DB

func CreateDB() error {
	_, _, err := DB.Run(ql.NewRWCtx(), `
			BEGIN TRANSACTION;
			CREATE TABLE properties (
				key string,
				value string
			);

			INSERT INTO 	properties
			VALUES 				("dbversion", "2");

			CREATE TABLE message (
				id 							string,
				receiveCount 		int,
				mailbox 				string,
				createdAt 			time,
				lastReceivedAt 	time,
				deployment 			string,
				deleted 				bool
			);

			CREATE TABLE mailbox (
				id  								string,
				createdAt 					time,
				version 						string,
				host 								string,
				lastCheckedInAt 		time
			);

			CREATE TABLE accessToken (
				mailbox 		string,
				token 			string,
				name 				string,
				fullAccess 	bool
			);

			CREATE TABLE deployment (
				id 						string,
				messageBody 	string,
				name 					string,
				deployedAt 		time,
				deployedBy 		string,
				totalMessages int,
				open 					bool,
				asset 				string
			);

			CREATE TABLE deploymentResponse (
				deployment 		string,
				mailbox 			string,
				response 			string,
				respondedAt 	time,
				isError 			bool
			);
			COMMIT;`)
	return err
}

func migrateDatabase() error {
	dbVersionStr, err := GetDBVersion()
	dbVersion, err := strconv.Atoi(dbVersionStr)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	switch {
	case dbVersion < 2:
		err := runMigration("2", `
			ALTER TABLE mailbox ADD version string;
			ALTER TABLE mailbox ADD host string;
		`)
		if err != nil {
			return err
		}
		fallthrough
	default:
	}
	return nil
}

func runMigration(version, sql string) error {
	log.Infof("Upgrading database to version %s", version)
	fullSql := `
		BEGIN TRANSACTION;
		` + sql + `
		UPDATE properties SET value = $1 WHERE key == "dbversion";
		COMMIT;`
	_, _, err := DB.Run(ql.NewRWCtx(), fullSql, version)
	return err
}

func OpenDB() {
	var err error
	shouldCreate := false
	directory, _ := osext.ExecutableFolder()
	path := filepath.Join(directory, "mailboxes.db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		shouldCreate = true
	}
	DB, err = ql.OpenFile(path, &ql.Options{CanCreate: true})
	if err != nil {
		fmt.Println("Could not open mailbox database.")
		os.Exit(-1)
	}
	if shouldCreate {
		err := CreateDB()
		if err != nil {
			panic(err)
		}
	}
	err = migrateDatabase()
	if err != nil {
		panic(err)
	}
}

func OpenMemDB() error {
	var err error
	DB, err = ql.OpenMem()
	return err
}

func GetDBVersion() (string, error) {
	var (
		dbVersion = ""
	)
	rss, _, err := DB.Run(ql.NewRWCtx(), `

	SELECT 	value
	FROM 		properties
	WHERE  	key == "dbversion"`)
	if err != nil {
		return dbVersion, err
	}
	rss[0].Do(false, func(data []interface{}) (bool, error) {
		dbVersion = data[0].(string)
		return false, nil
	})
	return dbVersion, err
}

func CloseDB() error {
	err := DB.Close()
	return err
}
