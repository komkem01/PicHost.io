package base

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

// Select Query
type QueryFunc func(*bun.SelectQuery) *bun.SelectQuery

// Update Query
type (
	ExecUpdateFunc func(*bun.UpdateQuery) *bun.UpdateQuery
	SetUpdateFunc  func(*bun.UpdateQuery) *bun.UpdateQuery
)

// DeleteQuery
type ExecDeleteFunc func(*bun.DeleteQuery) *bun.DeleteQuery

// TxQuery
type TxQueryFunc func(ctx context.Context, tx bun.Tx) error

// QueryInstant
type QueryInstant struct {
	DB *bun.DB
}

func NewInstant(db *bun.DB) *QueryInstant {
	return &QueryInstant{
		DB: db,
	}
}

func (s *QueryInstant) APIKey(ctx context.Context, apiKey string) error {
	return nil
}

func (s *QueryInstant) Exec(ctx context.Context, query string) error {
	_, err := s.DB.Exec(query)
	return err
}

func (s *QueryInstant) Insert(ctx context.Context, model any) error {
	_, err := s.DB.NewInsert().Model(model).Exec(ctx)
	return err
}

func (s *QueryInstant) InsertWithTableName(ctx context.Context, tableName string, model any) error {
	sql := s.DB.NewInsert().Model(model)
	if tableName != "" {
		sql = sql.ModelTableExpr(tableName)
	}
	_, err := sql.Exec(ctx)
	return err
}

func (s *QueryInstant) InsertWithIgnore(ctx context.Context, model any) error {
	_, err := s.DB.NewInsert().Model(model).Ignore().Exec(ctx)
	return err
}

func (s *QueryInstant) InsertWithIgnoreResult(ctx context.Context, model any) (sql.Result, error) {
	return s.DB.NewInsert().Model(model).Ignore().Exec(ctx)
}

func (s *QueryInstant) Delete(ctx context.Context, model any) error {
	_, err := s.DB.NewDelete().Model(model).WherePK().Exec(ctx)
	return err
}

func (s *QueryInstant) DeleteWithCondition(ctx context.Context, model any, fnQuery ExecDeleteFunc) error {
	selQ := s.DB.NewDelete().Model(model)
	if fnQuery != nil {
		selQ = fnQuery(selQ)
	}
	_, err := selQ.Exec(ctx)
	return err
}

func (s *QueryInstant) Update(ctx context.Context, model any, byID bool, cols ...string) error {
	q := s.DB.NewUpdate().Model(model).Column(cols...)

	if byID {
		q.Where(`id = ?id`)
	} else {
		q.WherePK()
	}

	_, err := q.Exec(ctx)
	return err
}

func (s *QueryInstant) UpdateWithCondition(ctx context.Context, model any, fnQuery ExecUpdateFunc, fnValue SetUpdateFunc) error {
	sqlQ := s.DB.NewUpdate().Model(model)
	if fnQuery != nil {
		sqlQ = fnQuery(sqlQ)
	}
	if fnValue != nil {
		sqlQ = fnValue(sqlQ)
	}
	_, err := sqlQ.Exec(ctx)
	return err
}

func (s *QueryInstant) GetBys(ctx context.Context, model any, fn QueryFunc) error {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}
	return selQ.Scan(ctx)
}

func (s *QueryInstant) CountBys(ctx context.Context, model any, fn QueryFunc) (int, error) {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}
	return selQ.Count(ctx)
}

func (s *QueryInstant) GetList(ctx context.Context, model any, req *RequestPaginate, allowSearchBy, allowOrderBy []string, fn QueryFunc) (any, *ResponsePaginate, error) {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}

	if allowSearchBy != nil {
		err := req.SetSearchBy(selQ, allowSearchBy)
		if err != nil {
			return nil, nil, err
		}
	}

	if allowOrderBy != nil {
		err := req.SetSortOrder(selQ, allowOrderBy)
		if err != nil {
			return nil, nil, err
		}
	}

	count, err := selQ.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	pag := &ResponsePaginate{
		Page:  req.GetPage(),
		Size:  req.GetSize(),
		Total: int64(count),
	}
	if count == 0 {
		return model, pag, nil
	}

	req.SetOffsetLimit(selQ)
	if err := selQ.Scan(ctx); err != nil {
		return nil, nil, err
	}

	return model, pag, nil
}

func (s *QueryInstant) GetListAll(ctx context.Context, model any, fn QueryFunc) (any, error) {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}
	if err := selQ.Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *QueryInstant) CustomQuery(ctx context.Context, model any, fn QueryFunc) (any, error) {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}
	if err := selQ.Scan(ctx); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *QueryInstant) Rows(ctx context.Context, model any, fn QueryFunc, limit int) (*sql.Rows, error) {
	selQ := s.DB.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}

	if limit != 0 {
		selQ.Limit(limit)
	}

	return selQ.Rows(ctx)
}

func (s *QueryInstant) ScanRows(ctx context.Context, rows *sql.Rows, model any) error {
	return s.DB.ScanRow(ctx, rows, model)
}

func (s *QueryInstant) RunInTx(ctx context.Context, bunTx TxQueryFunc, txOption *sql.TxOptions) error {
	return s.DB.RunInTx(ctx, txOption, bunTx)
}

func (s *QueryInstant) InitBunTx(bunTx bun.Tx) *BunTx {
	return &BunTx{bunTx}
}

type BunTx struct {
	bun.Tx
}

func (tx *BunTx) TxUpdateWithCondition(ctx context.Context, model any, fnQuery ExecUpdateFunc) (sql.Result, error) {
	selQ := tx.NewUpdate().Model(model)
	if fnQuery != nil {
		selQ = fnQuery(selQ)
	}
	return selQ.Exec(ctx)
}

func (tx *BunTx) TxInsert(ctx context.Context, tableName string, model any) error {
	sql := tx.NewInsert().Model(model)
	if tableName != "" {
		sql = sql.ModelTableExpr(tableName)
	}
	_, err := sql.Exec(ctx)
	return err
}

func (tx *BunTx) TxGetBys(ctx context.Context, model any, fn QueryFunc) error {
	selQ := tx.NewSelect().Model(model)
	if fn != nil {
		selQ = fn(selQ)
	}
	return selQ.Scan(ctx)
}
