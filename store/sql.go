package store

import (
	"database/sql"

	"github.com/go-jet/jet/v2/postgres"

	"github.com/connylabs/model-tracking/store/model-tracking/public/model"
	"github.com/connylabs/model-tracking/store/model-tracking/public/table"
)

type sqlStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) ModelTracking {
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
	db *sql.DB
}

func NewOrganizationsSQLStore(db *sql.DB) Organizations {
	return &organizationsSQLStore{db}
}

func (oss *organizationsSQLStore) Create(o *model.Organization) (*model.Organization, error) {
	var res model.Organization
	if err := table.Organization.INSERT(
		table.Organization.Name,
	).VALUES(
		o.Name,
	).RETURNING(
		table.Organization.AllColumns,
	).Query(oss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type modelsSQLStore struct {
	db           *sql.DB
	organization string
}

func NewModelsSQLStore(db *sql.DB, organization string) Models {
	return &modelsSQLStore{db, organization}
}

func (mss *modelsSQLStore) Create(m *model.Model) (*model.Model, error) {
	var o model.Organization
	if err := postgres.SELECT(
		table.Organization.ID,
	).FROM(
		table.Organization,
	).WHERE(
		table.Organization.Name.EQ(postgres.String(mss.organization)),
	).Query(mss.db, &o); err != nil {
		return nil, err
	}

	var res model.Model
	if err := table.Model.INSERT(
		table.Model.Name,
		table.Model.Organization,
	).VALUES(
		m.Name,
		o.ID,
	).RETURNING(
		table.Model.AllColumns,
	).Query(mss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (mss *modelsSQLStore) Get(name string) (*model.Model, error) {
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
	).Query(mss.db, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (mss *modelsSQLStore) List() ([]*model.Model, error) {
	var m []*model.Model
	if err := postgres.SELECT(
		table.Model.AllColumns,
	).FROM(
		table.Model.
			INNER_JOIN(table.Organization, table.Model.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(mss.organization))),
			),
	).Query(mss.db, &m); err != nil {
		return nil, err
	}

	return m, nil
}

type schemasSQLStore struct {
	db           *sql.DB
	organization string
}

func NewSchemasSQLStore(db *sql.DB, organization string) Schemas {
	return &schemasSQLStore{db, organization}
}

func (sss *schemasSQLStore) Create(s *model.Schema) (*model.Schema, error) {
	var o model.Organization
	if err := postgres.SELECT(
		table.Organization.ID,
	).FROM(
		table.Organization,
	).WHERE(
		table.Organization.Name.EQ(postgres.String(sss.organization)),
	).Query(sss.db, &o); err != nil {
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
	).Query(sss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (sss *schemasSQLStore) Get(name string) (*model.Schema, error) {
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
	).Query(sss.db, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sss *schemasSQLStore) GetByID(id int) (*model.Schema, error) {
	var s model.Schema
	if err := postgres.SELECT(
		table.Schema.AllColumns,
	).FROM(
		table.Schema,
	).WHERE(
		table.Schema.ID.EQ(postgres.Int(int64((id)))),
	).Query(sss.db, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func (sss *schemasSQLStore) List() ([]*model.Schema, error) {
	var s []*model.Schema
	if err := postgres.SELECT(
		table.Schema.AllColumns,
	).FROM(
		table.Schema.
			INNER_JOIN(table.Organization, table.Schema.Organization.EQ(table.Organization.ID).
				AND(table.Organization.Name.EQ(postgres.String(sss.organization))),
			),
	).Query(sss.db, &s); err != nil {
		return nil, err
	}

	return s, nil
}

type versionsSQLStore struct {
	db           *sql.DB
	organization string
	model        string
}

func NewVersionsSQLStore(db *sql.DB, organization, model string) Versions {
	return &versionsSQLStore{db, organization, model}
}

func (vss *versionsSQLStore) Create(v *model.Version) (*model.Version, error) {
	m, err := (&modelsSQLStore{db: vss.db, organization: vss.organization}).Get(vss.model)
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
	).Query(vss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (vss *versionsSQLStore) Get(name string) (*model.Version, error) {
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
	).Query(vss.db, &v); err != nil {
		return nil, err
	}

	return &v, nil
}

func (vss *versionsSQLStore) List() ([]*model.Version, error) {
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
	).Query(vss.db, &v); err != nil {
		return nil, err
	}

	return v, nil
}

type resultsSQLStore struct {
	db           *sql.DB
	organization string
	model        string
	version      string
}

func NewResultsSQLStore(db *sql.DB, organization, model, version string) Results {
	return &resultsSQLStore{db, organization, model, version}
}

func (rss *resultsSQLStore) Create(r *model.Result) (*model.Result, error) {
	v, err := (&versionsSQLStore{db: rss.db, organization: rss.organization, model: rss.model}).Get(rss.version)
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
	).Query(rss.db, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (rss *resultsSQLStore) Get(id int) (*model.Result, error) {
	var r model.Result
	if err := postgres.SELECT(
		table.Result.AllColumns,
	).FROM(
		table.Result,
	).WHERE(
		table.Result.ID.EQ(postgres.Int(int64(id))),
	).Query(rss.db, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (rss *resultsSQLStore) List() ([]*model.Result, error) {
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
	).Query(rss.db, &r); err != nil {
		return nil, err
	}

	return r, nil
}
