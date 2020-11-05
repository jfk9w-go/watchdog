package server

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jfk9w-go/watchdog/client"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var DatabaseTable = goqu.T("alert")

type Database struct {
	*goqu.Database
}

func NewDatabase(driver, datasource string) (*Database, error) {
	db, err := sql.Open(driver, datasource)
	if err != nil {
		return nil, errors.Wrapf(err, "open %s", datasource)
	}
	return &Database{Database: goqu.New(driver, db)}, nil
}

func (db *Database) Init(ctx context.Context) error {
	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
	  service_id VARCHAR(63) NOT NULL,
      hostname VARCHAR(63) NOT NULL,
	  start TIMESTAMP NOT NULL,
	  time TIMESTAMP NOT NULL,
	  state SMALLINT NOT NULL,
	  error VARCHAR(511),
	  UNIQUE(service_id, hostname, start))`, DatabaseTable.GetTable())
	_, err := db.ExecContext(ctx, query)
	return err
}

var AlertColumns = []interface{}{"service_id", "hostname", "start", "time", "state", "error"}

func (db *Database) Alert(ctx context.Context, alert client.Alert) error {
	query, _, _ := db.Insert(DatabaseTable).
		Cols(AlertColumns...).
		Vals([]interface{}{
			alert.ServiceID, alert.Hostname,
			alert.Start.In(time.UTC), alert.Time.In(time.UTC),
			alert.State, alert.Error,
		}).
		OnConflict(goqu.DoUpdate(
			"service_id, hostname, start",
			map[string]interface{}{
				"time":  alert.Time,
				"state": alert.State,
				"error": alert.Error,
			})).
		ToSQL()

	_, err := db.ExecContext(ctx, query)
	return err
}

func (db *Database) Stats(ctx context.Context, id string, node string, last int) ([]client.Alert, error) {
	runs := make([]client.Alert, 0)
	return runs, db.Select(client.Alert{}).
		From(DatabaseTable).
		Where(goqu.And(
			goqu.C("service_id").Eq(id),
			goqu.C("hostname").Eq(node),
			goqu.C("state").Gt(client.Active))).
		Order(goqu.C("time").Desc()).
		Limit(uint(last)).
		ScanStructsContext(ctx, &runs)
}

func (db *Database) Close() error {
	return db.Db.(*sql.DB).Close()
}
