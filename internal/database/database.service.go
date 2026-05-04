package database

import (
	"context"
	"fmt"
	"sync"

	dto "mcop/internal/database/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jinzhu/copier"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	// "github.com/uptrace/bun/extra/bundebug"
)

type DatabaseService struct {
	dbMap  map[string]*bun.DB
	dbLock sync.RWMutex
}

var defaultDBConfig = dto.Option{
	TimeZone: "Asia/Bangkok",
}

func newService(opts map[string]*dto.Option) *DatabaseService {
	service := &DatabaseService{
		dbMap: make(map[string]*bun.DB),
	}
	if err := service.Register(context.Background(), opts); err != nil {
		panic(err)
	}
	return service
}

func (dbs *DatabaseService) DB(name ...string) *bun.DB {
	dbs.dbLock.RLock()
	defer dbs.dbLock.RUnlock()

	if dbs.dbMap == nil {
		panic("database not initialized")
	}

	if len(name) == 0 {
		return dbs.dbMap[""]
	}

	db, ok := dbs.dbMap[name[0]]
	if !ok {
		panic("database not initialized")
	}
	return db
}

func (dbs *DatabaseService) Register(ctx context.Context, opts map[string]*dto.Option) error {
	dbs.dbLock.Lock()
	defer dbs.dbLock.Unlock()
	for key, opt := range opts {
		optDef := withDBDefaultConf(opt)
		config, err := pgx.ParseConfig(optDef.Dsn)
		if err != nil {
			return err
		}
		sqldb := stdlib.OpenDB(*config)
		dbs.dbMap[key] = bun.NewDB(sqldb, pgdialect.New())
		if err := dbs.dbMap[key].PingContext(ctx); err != nil {
			return err
		}
		// dbs.dbMap[key].AddQueryHook(bundebug.NewQueryHook())
	}
	return nil
}

func (dbs *DatabaseService) Close(ctx context.Context, name string) error {
	dbs.dbLock.Lock()
	defer dbs.dbLock.Unlock()
	if dbs.dbMap == nil {
		return nil
	}
	db, ok := dbs.dbMap[name]
	if !ok {
		return nil
	}
	if err := db.Close(); err != nil {
		return err
	}
	delete(dbs.dbMap, name)
	return nil
}

func (dbs *DatabaseService) close(ctx context.Context) error {
	dbs.dbLock.Lock()
	defer dbs.dbLock.Unlock()
	for _, db := range dbs.dbMap {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

func withDBDefaultConf(opt *dto.Option) *dto.Option {
	rOpt := defaultDBConfig
	copier.CopyWithOption(&rOpt, opt, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if rOpt.Dsn == "" {
		rOpt.Dsn = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable TimeZone=%s", rOpt.Host, rOpt.Port, rOpt.Database, rOpt.Username, rOpt.Password, rOpt.TimeZone)
	}
	return &rOpt
}
