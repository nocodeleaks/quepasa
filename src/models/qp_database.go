package models

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"path/filepath"
	"runtime"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	migrate "github.com/joncalhoun/migrate"
	library "github.com/nocodeleaks/quepasa/library"
	whatsmeow "github.com/nocodeleaks/quepasa/whatsmeow"
	log "github.com/sirupsen/logrus"
)

type QpDatabase struct {
	Parameters  library.DatabaseParameters `json:"parameters,omitempty"`
	Connection  *sqlx.DB
	Users       QpDataUsersInterface
	Servers     QpDataServersInterface
	Dispatching QpDataDispatchingInterface
}

var (
	Sync       sync.Once // Objeto de sinaleiro para garantir uma única chamada em todo o andamento do programa
	Connection *sqlx.DB
)

var dbParameters = library.DatabaseParameters{
	Driver:   "sqlite3",
	DataBase: "quepasa",
}

// GetDB returns a database connection for the given
// database environment variables
func GetDB() *sqlx.DB {
	Sync.Do(func() {

		// generates the relative connection string
		connectionString := dbParameters.GetConnectionString()

		// Tenta realizar a conexão
		dbconn, err := sqlx.Connect(dbParameters.Driver, connectionString)
		if err != nil {
			log.Fatalf("error at database connection: %s, msg: %s", dbParameters.Driver, err.Error())
			return
		}

		dbconn.DB.SetMaxIdleConns(500)
		dbconn.DB.SetMaxOpenConns(1000)
		dbconn.DB.SetConnMaxLifetime(30 * time.Second)

		// Definindo uma única conexão para todo o sistema
		Connection = dbconn
	})
	return Connection
}

func GetDatabase() *QpDatabase {
	db := GetDB()

	var iusers = QpDataUserSql{db}
	var iservers = QpDataServerSql{db}
	var idispatching = QpDataServerDispatchingSql{db}

	return &QpDatabase{
		dbParameters,
		db,
		iusers,
		iservers,
		idispatching}
}

// MigrateToLatest updates the database to the latest schema
func MigrateToLatest(logentry *log.Entry) (err error) {
	if !ENV.Migrate() {
		return
	}

	fullPath := ENV.MigrationPath()

	logentry.Info("migrating database (if necessary)")
	if len(fullPath) == 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			logentry.Info("migrating database on windows operational system")

			// windows ===================
			leadingWindowsUnit, _ := filepath.Rel("z:\\", workDir)
			migrationsDir := filepath.Join(leadingWindowsUnit, "migrations")
			fullPath = fmt.Sprintf("/%s", strings.ReplaceAll(migrationsDir, "\\", "/"))
		} else {
			// linux ===================
			migrationsDir := filepath.Join(workDir, "migrations")
			fullPath = fmt.Sprintf("file://%s/", strings.Trim(migrationsDir, "/"))
		}
	}

	logentry.Debugf("full path database: %s", fullPath)

	migrations := Migrations(fullPath)
	db := GetDB().DB
	migrator := &QpMigrator{
		Migrations: migrations,
		LogStruct:  library.LogStruct{LogEntry: logentry},
	}
	err = migrator.Migrate(db, dbParameters.Driver)
	if err != nil {
		logentry.Fatal(err)
	}

	logentry.Debug("migrating finished")
	return nil
}

func Migrations(fullPath string) (migrations []migrate.SqlxMigration) {
	log.Debugf("migrating files from: %s", fullPath)
	files, err := os.ReadDir(fullPath)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the file specified") {
			log.Warnf("no migrations found at: %s", fullPath)
		} else {
			log.Fatal(err)
		}
	}

	log.Debug("migrating creating array with definitions")
	confMap := make(map[string]*QpMigration)

	for _, file := range files {
		info := file.Name()
		dotSplitted := strings.Split(info, ".")      // file name splitted by dots
		extension := dotSplitted[len(dotSplitted)-1] // file extension
		if extension == "sql" {
			id := strings.Split(info, "_")[0]

			title := strings.TrimPrefix(dotSplitted[0], id+"_")
			status := dotSplitted[1]
			filepath := fullPath + "/" + info
			if v, ok := confMap[id]; ok {
				switch status {
				case "up":
					v.FileUp = filepath
				case "down":
					v.FileDown = filepath
				default:
					// unknown status - ignore
				}
			} else {
				switch status {
				case "up":
					confMap[id] = &QpMigration{Id: id, Title: title, FileUp: filepath}
				case "down":
					confMap[id] = &QpMigration{Id: id, Title: title, FileDown: filepath}
				default:
					// unknown status - ignore
				}
			}
		}
	}

	migrations = append(migrations, GetBase())

	for _, migration := range confMap {
		migrations = append(migrations, migration.ToSqlxMigration())
	}

	// ordering
	sort.SliceStable(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return
}

// Provides the first migration
func GetBase() migrate.SqlxMigration {
	migration := migrate.SqlxQueryMigration("1", `
	CREATE TABLE IF NOT EXISTS "users" (
		"username" CHAR (255) PRIMARY KEY NOT NULL,
		"password" VARCHAR (255) NOT NULL,
		"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	  );
	  
	  CREATE TABLE IF NOT EXISTS "servers" (
		"token" CHAR (100) PRIMARY KEY UNIQUE NOT NULL,
		"wid" VARCHAR (255) UNIQUE NOT NULL,
		"verified" BOOLEAN NOT NULL DEFAULT FALSE,
		"devel" BOOLEAN NOT NULL DEFAULT FALSE,
		"groups" INT(1) NOT NULL DEFAULT 0,
  		"broadcasts" INT(1) NOT NULL DEFAULT 0,
  		"readreceipts" INT(1) NOT NULL DEFAULT 0,
  		"calls" INT(1) NOT NULL DEFAULT 0,
		"user" CHAR (255) DEFAULT NULL REFERENCES "users"("username"),
		"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	  );
	  
	  CREATE TABLE IF NOT EXISTS "dispatching" (
		"context" CHAR (100) NOT NULL REFERENCES "servers"("token"),
		"connection_string" VARCHAR (255) NOT NULL,
		"type" VARCHAR (50) NOT NULL DEFAULT 'webhook',
		"forwardinternal" BOOLEAN NOT NULL DEFAULT FALSE,
		"trackid" VARCHAR (100) NOT NULL DEFAULT '',
		"readreceipts" INT(1) NOT NULL DEFAULT 0,
  		"groups" INT(1) NOT NULL DEFAULT 0,
  		"broadcasts" INT(1) NOT NULL DEFAULT 0,
		"extra" BLOB DEFAULT NULL,
		"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT "dispatching_pkey" PRIMARY KEY ("context", "connection_string")
	  );
	  
	  INSERT OR REPLACE INTO migrations (id) VALUES
	  ('202207131700'),
	  ('202209281840'),
	  ('202303011900'),
	  ('202402291556'),
	  ('202403021242'),
	  ('202403141920'),
	  ('202512151400');
	  `, "")
	return migration
}

type QpMigrator struct {
	Migrations []migrate.SqlxMigration
	library.LogStruct
}

func (source *QpMigrator) Printf(format string, args ...interface{}) (int, error) {
	format = strings.ToLower(format)
	format = strings.ReplaceAll(format, "\n", "")

	logentry := source.GetLogger()
	logentry.Debugf(format, args)

	if len(args) > 0 && strings.Contains(format, "running") {
		str := fmt.Sprintf("%s", args[0])
		Running = append(Running, str)
	}

	return len([]byte(format)), nil
}

func (source *QpMigrator) Migrate(sqlDB *sql.DB, dialect string) error {
	migrator := &migrate.Sqlx{
		Printf:     source.Printf,
		Migrations: source.Migrations,
	}

	err := migrator.Migrate(sqlDB, dialect)
	if err != nil {
		rbErr := migrator.Rollback(sqlDB, dialect)
		if rbErr != nil {
			return rbErr
		}
		return err
	}

	return nil
}

var Running []string

var MigrationHandlers = map[string]func(string){
	"202303011900": MigrationHandler_202303011900,
}

func MigrationHandler_202303011900(id string) {
	log.Infof("running migration handler for: %s", id)
	db := GetDatabase()
	servers := db.Servers.FindAll()
	for _, server := range servers {
		oldWid := server.Wid
		if strings.HasSuffix(oldWid, "@migrated") {
			phone := library.GetPhoneByWId(oldWid)
			store, err := whatsmeow.WhatsmeowService.GetStoreForMigrated(phone)
			if err != nil {
				log.Warnf("error at getting store for phone: %s, cause: %s", phone, err.Error())
				continue
			}

			server.Wid = store.ID.String()
			err = db.Servers.Update(server)
			if err != nil {
				log.Fatalf("error at update server: %s", err.Error())
			}

			log.Infof("wid updated with success: %s", server.Token)

			// Removido bloco legado que usava db.Webhooks (deprecated)
			// Todos os webhooks e RabbitMQ agora são tratados via Dispatching
		}
	}
}
