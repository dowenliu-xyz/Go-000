package di

import (
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/ent"
	"github.com/facebook/ent/dialect"
	"github.com/facebook/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
)

func NewClient(database Database) (*ent.Client, func(), error) {
	drv, err := sql.Open(dialect.MySQL, database.DSN)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "打开数据库连接失败, dsn: %s", database.DSN)
	}
	db := drv.DB()
	if database.MaxIdleConns != nil {
		db.SetMaxIdleConns(*database.MaxIdleConns)
	}
	if database.MaxOpenConns != nil {
		db.SetMaxOpenConns(*database.MaxOpenConns)
	}
	if database.ConnMaxIdleTimeSec != nil {
		db.SetConnMaxIdleTime(time.Duration(*database.ConnMaxIdleTimeSec) * time.Second)
	}
	client := ent.NewClient(ent.Driver(drv))
	return client, func() {
		_ = client.Close()
	}, nil
}
