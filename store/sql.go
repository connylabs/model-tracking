package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"github.com/connylabs/model-tracking/store/model-tracking/public/model"
	"github.com/connylabs/model-tracking/store/model-tracking/public/table"
)

type tx struct {
	*sql.Tx
	managedExternally bool
}

func (t *tx) Commit() error {
	if t.managedExternally {
		return nil
	}

	return t.Tx.Commit()
}

func (t *tx) Rollback() error {
	if t.managedExternally {
		return nil
	}

	return t.Tx.Rollback()
}

type txable struct {
	qrm.DB
	called atomic.Bool
}

func (t *txable) Begin() (*tx, error) {
	return t.BeginTx(context.Background(), nil)
}

func (t *txable) BeginTx(ctx context.Context, opts *sql.TxOptions) (*tx, error) {
	switch v := t.DB.(type) {
	case *sql.Tx:
		if t.called.CompareAndSwap(false, true) {
			return &tx{v, true}, nil
		}
		return nil, errors.New("the underlying transaction has already been used")
	case *sql.DB:
		t, err := v.BeginTx(ctx, nil)
		return &tx{t, false}, err
	case *tx:
		return &tx{v.Tx, true}, nil
	default:
		return nil, fmt.Errorf("cannot create a transaction from %T", v)
	}
}

func newTxable(db qrm.DB) *txable {
	return &txable{db, atomic.Bool{}}
}

type sqlStore struct {
	db qrm.DB
}

func NewSQLStore(db qrm.DB) ModelTracking {
	return &sqlStore{db}
}

func (ss *sqlStore) Organizations() Organizations {
	return NewOrganizationsSQLStore(ss.db)
}

func (ss *sqlStore) Models(organization string) Models {
	return NewModelsSQLStore(ss.db, organization)
}

func (ss *sqlStore) Schemas(organization string) Schemas {
	return NewSchemasSQLStore(ss.db, organization)
}

func (ss *sqlStore) Versions(organization, model string) Versions {
	return NewVersionsSQLStore(ss.db, organization, model)
}

func (ss *sqlStore) Results(organization, model, version string) Results {
	return NewResultsSQLStore(ss.db, organization, model, version)
}

type organizationsSQLStore struct {
	db qrm.DB
}

func NewOrganizationsSQLStore(db qrm.DB) Organizations {
	return &organizationsSQLStore{db}
}

func (oss *organizationsSQLStore) Create(ctx context.Context, o *model.Organization) (*model.Organization, error) {
	var res model.Organization
	if err := table.Organization.INSERT(
		table.Organization.Name,
	).VALUES(
		o.Name,
	).RETURNING(
		table.Organization.AllColumns,
	).QueryContext(ctx, oss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type modelsSQLStore struct {
	db           qrm.DB
	organization string
}

func NewModelsSQLStore(db qrm.DB, organization string) Models {
	return &modelsSQLStore{db, organization}
}

func (mss *modelsSQLStore) Create(ctx context.Context, m *model.Model) (*model.Model, error) {
	tx, err := newTxable(mss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var o model.Organization
	if err := postgres.SELECT(
		table.Organization.ID,
	).FROM(
		table.Organization,
	).WHERE(
		table.Organization.Name.EQ(postgres.String(mss.organization)),
	).QueryContext(ctx, tx, &o); err != nil {
		return nil, err
	}

	if m.DefaultSchema != nil {
		if _, err := NewSchemasSQLStore(tx, mss.organization).GetByID(ctx, int(*m.DefaultSchema)); err != nil {
			return nil, fmt.Errorf("could not find schema: %w", err)
		}
	}

	var res model.Model
	if err := table.Model.INSERT(
		table.Model.Name,
		table.Model.Organization,
		table.Model.DefaultSchema,
	).VALUES(
		m.Name,
		o.ID,
		m.DefaultSchema,
	).RETURNING(
		table.Model.AllColumns,
	).QueryContext(ctx, tx, &res); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (mss *modelsSQLStore) Update(ctx context.Context, m *model.Model) (*model.Model, error) {
	tx, err := newTxable(mss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	// Ensure this model exists in the desired organization.
	if _, err := NewModelsSQLStore(tx, mss.organization).Get(ctx, m.Name); err != nil {
		return nil, err
	}

	if m.DefaultSchema != nil {
		if _, err := NewSchemasSQLStore(tx, mss.organization).GetByID(ctx, int(*m.DefaultSchema)); err != nil {
			return nil, fmt.Errorf("could not find schema: %w", err)
		}
	}
	var res model.Model
	if err := table.Model.UPDATE(
		table.Model.DefaultSchema,
	).SET(
		m.DefaultSchema,
	).WHERE(
		table.Model.Name.EQ(postgres.String(m.Name)),
	).RETURNING(
		table.Model.AllColumns,
	).QueryContext(ctx, tx, &res); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (mss *modelsSQLStore) Get(ctx context.Context, name string) (*model.Model, error) {
	var m model.Model
	if err := postgres.SELECT(
		table.Model.AllColumns,
	).FROM(
		table.Model.
			INNER_JOIN(table.Organization, table.Model.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(mss.organization))),
			),
	).WHERE(
		table.Model.Name.EQ(postgres.String(name)),
	).QueryContext(ctx, mss.db, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (mss *modelsSQLStore) List(ctx context.Context) ([]*model.Model, error) {
	var m []*model.Model
	if err := postgres.SELECT(
		table.Model.AllColumns,
	).FROM(
		table.Model.
			INNER_JOIN(table.Organization, table.Model.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(mss.organization))),
			),
	).QueryContext(ctx, mss.db, &m); err != nil {
		return nil, err
	}

	return m, nil
}

type schemasSQLStore struct {
	db           qrm.DB
	organization string
}

func NewSchemasSQLStore(db qrm.DB, organization string) Schemas {
	return &schemasSQLStore{db, organization}
}

func (sss *schemasSQLStore) Create(ctx context.Context, s *model.Schema) (*model.Schema, error) {
	tx, err := newTxable(sss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	var o model.Organization
	if err := postgres.SELECT(
		table.Organization.ID,
	).FROM(
		table.Organization,
	).WHERE(
		table.Organization.Name.EQ(postgres.String(sss.organization)),
	).QueryContext(ctx, tx, &o); err != nil {
		return nil, err
	}

	var res model.Schema
	if err := table.Schema.INSERT(
		table.Schema.Name,
		table.Schema.Organization,
		table.Schema.Input,
		table.Schema.Output,
	).VALUES(
		s.Name,
		o.ID,
		s.Input,
		s.Output,
	).RETURNING(
		table.Schema.AllColumns,
	).QueryContext(ctx, tx, &res); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (sss *schemasSQLStore) Get(ctx context.Context, name string) (*model.Schema, error) {
	var s model.Schema
	if err := postgres.SELECT(
		table.Schema.AllColumns,
	).FROM(
		table.Schema.
			INNER_JOIN(table.Organization, table.Schema.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(sss.organization))),
			),
	).WHERE(
		table.Schema.Name.EQ(postgres.String(name)),
	).QueryContext(ctx, sss.db, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sss *schemasSQLStore) GetByID(ctx context.Context, id int) (*model.Schema, error) {
	var s model.Schema
	if err := postgres.SELECT(
		table.Schema.AllColumns,
	).FROM(
		table.Schema,
	).WHERE(
		table.Schema.ID.EQ(postgres.Int(int64((id)))),
	).QueryContext(ctx, sss.db, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sss *schemasSQLStore) List(ctx context.Context) ([]*model.Schema, error) {
	var s []*model.Schema
	if err := postgres.SELECT(
		table.Schema.AllColumns,
	).FROM(
		table.Schema.
			INNER_JOIN(table.Organization, table.Schema.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(sss.organization))),
			),
	).QueryContext(ctx, sss.db, &s); err != nil {
		return nil, err
	}

	return s, nil
}

type versionsSQLStore struct {
	db           qrm.DB
	organization string
	model        string
}

func NewVersionsSQLStore(db qrm.DB, organization, model string) Versions {
	return &versionsSQLStore{db, organization, model}
}

func (vss *versionsSQLStore) Create(ctx context.Context, v *model.Version) (*model.Version, error) {
	tx, err := newTxable(vss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	m, err := NewModelsSQLStore(tx, vss.organization).Get(ctx, vss.model)
	if err != nil {
		return nil, err
	}

	var res model.Version
	if err := table.Version.INSERT(
		table.Version.Name,
		table.Version.Model,
		table.Version.Organization,
		table.Version.Schema,
	).VALUES(
		v.Name,
		m.ID,
		m.Organization,
		v.Schema,
	).RETURNING(
		table.Version.AllColumns,
	).QueryContext(ctx, tx, &res); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &res, nil
}

func (vss *versionsSQLStore) Get(ctx context.Context, name string) (*model.Version, error) {
	var v model.Version
	if err := postgres.SELECT(
		table.Version.AllColumns,
	).FROM(
		table.Version.
			INNER_JOIN(table.Organization, table.Version.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(vss.organization))),
			).
			INNER_JOIN(table.Model, table.Version.Model.EQ(table.Model.ID).
				AND(table.Model.Name.EQ(postgres.String(vss.model))),
			),
	).WHERE(
		table.Version.Name.EQ(postgres.String(name)),
	).QueryContext(ctx, vss.db, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (vss *versionsSQLStore) GetOrCreate(ctx context.Context, name string) (*model.Version, error) {
	tx, err := newTxable(vss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	v, err := NewVersionsSQLStore(tx, vss.organization, vss.model).Get(ctx, name)
	if err == nil {
		return v, nil
	}
	if !errors.Is(err, qrm.ErrNoRows) {
		println("WAS NOT A NO RESULTS ERROR")
		return nil, err
	}

	println("WAS A NO RESULTS ERROR")
	m, err := NewModelsSQLStore(tx, vss.organization).Get(ctx, vss.model)
	if err != nil {
		println("DID NOT FIND A MODEL")
		return nil, err
	}

	if m.DefaultSchema == nil {
		return nil, qrm.ErrNoRows
	}

	v, err = NewVersionsSQLStore(tx, vss.organization, vss.model).Create(ctx, &model.Version{Name: name, Model: m.ID, Organization: m.Organization, Schema: *m.DefaultSchema})
	if err != nil {
		println("COULD NOT CREAT E THE VERSION")
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return v, nil
}

func (vss *versionsSQLStore) List(ctx context.Context) ([]*model.Version, error) {
	var v []*model.Version
	if err := postgres.SELECT(
		table.Version.AllColumns,
	).FROM(
		table.Version.
			INNER_JOIN(table.Organization, table.Version.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(vss.organization))),
			).
			INNER_JOIN(table.Model, table.Version.Model.EQ(table.Model.ID).
				AND(table.Model.Name.EQ(postgres.String(vss.model))),
			),
	).QueryContext(ctx, vss.db, &v); err != nil {
		return nil, err
	}

	return v, nil
}

type resultsSQLStore struct {
	db           qrm.DB
	organization string
	model        string
	version      string
}

func NewResultsSQLStore(db qrm.DB, organization, model, version string) Results {
	return &resultsSQLStore{db, organization, model, version}
}

func (rss *resultsSQLStore) Create(ctx context.Context, r *model.Result) (*model.Result, error) {
	tx, err := newTxable(rss.db).BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	v, err := NewVersionsSQLStore(tx, rss.organization, rss.model).Get(ctx, rss.version)
	if err != nil {
		return nil, err
	}

	var res model.Result
	if err := table.Result.INSERT(
		table.Result.Model,
		table.Result.Organization,
		table.Result.Version,
		table.Result.Input,
		table.Result.Output,
		table.Result.TrueOutput,
		table.Result.Time,
	).VALUES(
		v.Model,
		v.Organization,
		v.ID,
		r.Input,
		r.Output,
		r.TrueOutput,
		r.Time,
	).RETURNING(
		table.Result.AllColumns,
	).QueryContext(ctx, tx, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (rss *resultsSQLStore) Get(ctx context.Context, id int) (*model.Result, error) {
	var r model.Result
	if err := postgres.SELECT(
		table.Result.AllColumns,
	).FROM(
		table.Result,
	).WHERE(
		table.Result.ID.EQ(postgres.Int(int64(id))),
	).QueryContext(ctx, rss.db, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (rss *resultsSQLStore) List(ctx context.Context) ([]*model.Result, error) {
	var r []*model.Result
	if err := postgres.SELECT(
		table.Result.AllColumns,
	).FROM(
		table.Result.
			INNER_JOIN(table.Organization, table.Result.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(rss.organization))),
			).
			INNER_JOIN(table.Model, table.Result.Model.EQ(table.Model.ID).
				AND(table.Model.Name.EQ(postgres.String(rss.model))),
			).
			INNER_JOIN(table.Version, table.Result.Version.EQ(table.Version.ID).
				AND(table.Version.Name.EQ(postgres.String(rss.version))),
			),
	).QueryContext(ctx, rss.db, &r); err != nil {
		return nil, err
	}

	return r, nil
}
