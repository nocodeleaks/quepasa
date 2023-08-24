package models

import (
	"database/sql"
	"fmt"
	"io/ioutil"
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
	Config     QpDatabaseConfig `json:"config,omitempty"`
	Connection *sqlx.DB
	Users      QpDataUsersInterface
	Servers    QpDataServersInterface
	Webhooks   QpDataWebhooksInterface
}

var (
	Sync       sync.Once // Objeto de sinaleiro para garantir uma única chamada em todo o andamento do programa
	Connection *sqlx.DB
)

// GetDB returns a database connection for the given
// database environment variables
func GetDB() *sqlx.DB {
	Sync.Do(func() {
		config := GetDBConfig()

		// Tenta realizar a conexão
		dbconn, err := sqlx.Connect(config.Driver, config.GetConnectionString())
		if err != nil {
			log.Fatalf("error at database connection: %s, msg: %s", config.Driver, err.Error())
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
	config := GetDBConfig()
	var iusers = QpDataUserSql{db}
	var iwebhooks = QpDataServerWebhookSql{db}
	var iservers = QpDataServerSql{db}

	return &QpDatabase{
		config,
		db,
		iusers,
		iservers,
		iwebhooks}
}

func GetDBConfig() QpDatabaseConfig {
	config := QpDatabaseConfig{}

	config.Driver = os.Getenv("DBDRIVER")
	if len(config.Driver) == 0 {
		config.Driver = "sqlite3"
	}

	config.Host = os.Getenv("DBHOST")
	config.DataBase = os.Getenv("DBDATABASE")
	config.Port = os.Getenv("DBPORT")
	config.User = os.Getenv("DBUSER")
	config.Password = os.Getenv("DBPASSWORD")
	config.SSL = os.Getenv("DBSSLMODE")
	return config
}

// MigrateToLatest updates the database to the latest schema
func MigrateToLatest() (err error) {
	if !ENV.Migrate() {
		return
	}

	fullPath := ENV.MigrationPath()

	log.Info("migrating database (if necessary)")
	if len(fullPath) == 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			log.Info("migrating database on windows operational system")

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

	log.Debugf("fullpath database: %s", fullPath)

	migrations := Migrations(fullPath)
	config := GetDBConfig()
	db := GetDB().DB
	migrator := &QpMigrator{Migrations: migrations}
	err = migrator.Migrate(db, config.Driver)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("migrating finished")
	return nil
}

func Migrations(fullPath string) (migrations []migrate.SqlxMigration) {
	log.Debugf("migrating files from: %s", fullPath)
	files, err := ioutil.ReadDir(fullPath)
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
				if status == "up" {
					v.FileUp = filepath
				} else if status == "down" {
					v.FileDown = filepath
				}
			} else {
				if status == "up" {
					confMap[id] = &QpMigration{Id: id, Title: title, FileUp: filepath}
				} else if status == "down" {
					confMap[id] = &QpMigration{Id: id, Title: title, FileDown: filepath}
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
		"handlegroups" BOOLEAN NOT NULL DEFAULT TRUE,
		"handlebroadcast" BOOLEAN NOT NULL DEFAULT FALSE,
		"user" CHAR (36) DEFAULT NULL REFERENCES "users"("username"),
		"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	  );
	  
	  CREATE TABLE IF NOT EXISTS "webhooks" (
		"context" CHAR (255) NOT NULL REFERENCES "servers"("token"),
		"url" VARCHAR (255) NOT NULL,
		"forwardinternal" BOOLEAN NOT NULL DEFAULT FALSE,
		"trackid" VARCHAR (100) NOT NULL DEFAULT '',
		"extra" BLOB DEFAULT NULL,
		"timestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT "webhooks_pkey" PRIMARY KEY ("context", "url")
	  );
	  
	  INSERT OR REPLACE INTO migrations (id) VALUES
	  ('202207131700'),
	  ('202209281840'),
	  ('202303011900')
	  ;
	  `, "")
	return migration
}

type QpMigrator struct {
	Migrations []migrate.SqlxMigration
}

func (source *QpMigrator) Printf(format string, args ...interface{}) (int, error) {
	format = strings.ToLower(format)
	format = strings.ReplaceAll(format, "\n", "")
	log.Debugf(format, args)

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
		oldWid := server.WId
		if strings.HasSuffix(oldWid, "@migrated") {
			phone := library.GetPhoneByWId(oldWid)
			store, err := whatsmeow.WhatsmeowService.GetStoreForMigrated(phone)
			if err != nil {
				log.Warnf("error at getting store for phone: %s, cause: %s", phone, err.Error())
				continue
			}

			server.WId = store.ID.String()
			err = db.Servers.Update(server)
			if err != nil {
				log.Fatalf("error at update server: %s", err.Error())
			}

			log.Infof("wid updated with success: %s", server.Token)

			webhooks, err := db.Webhooks.FindAll(oldWid)
			if err != nil {
				log.Fatalf("cant get webhook from database")
			}

			for _, webhook := range webhooks {
				err = db.Webhooks.UpdateContext(webhook, server.Token)
				if err != nil {
					log.Fatalf("cant update webhook from database")
				}

				log.Infof("webhook updated with success for: %s, url: %s", webhook.Context, webhook.Url)
			}
		}
	}
}
