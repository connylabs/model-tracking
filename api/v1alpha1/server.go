package v1alpha1

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/xeipuuv/gojsonschema"

	"github.com/connylabs/model-tracking/store"
	"github.com/connylabs/model-tracking/store/model-tracking/public/model"
)

func httpError(logger log.Logger) func(w http.ResponseWriter, m string, code int) {
	hj := httpJSON(logger)
	return func(w http.ResponseWriter, m string, code int) {
		response := Error{
			Code:  code,
			Error: m,
		}
		hj(w, response, code)
	}
}

func httpJSON(logger log.Logger) func(w http.ResponseWriter, response interface{}, code int) {
	return func(w http.ResponseWriter, response interface{}, code int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			level.Error(logger).Log("msg", "failed to write response", "err", err.Error())
		}
	}
}

type server struct {
	store     store.ModelTracking
	logger    log.Logger
	httpError func(w http.ResponseWriter, m string, code int)
	httpJSON  func(w http.ResponseWriter, response interface{}, code int)
}

func NewServer(store store.ModelTracking, logger log.Logger) ServerInterface {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return &server{
		store:     store,
		logger:    logger,
		httpError: httpError(logger),
		httpJSON:  httpJSON(logger),
	}
}

func (s *server) OrganizationsCreate(w http.ResponseWriter, r *http.Request) {
	body := new(OrganizationsCreateJSONBody)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(body); err != nil {
		if errors.Is(err, (*json.UnmarshalTypeError)(nil)) {
			s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	o, err := s.store.Organizations().Create(&model.Organization{Name: body.Name})
	if err != nil {
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Organization{
		ID:      int(o.ID),
		Name:    o.Name,
		Created: *o.Created,
		Updated: *o.Updated,
	}, http.StatusCreated)
}

func (s *server) ModelsListForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization) {
	ms, err := s.store.Models(organization).List()
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	models := make([]*Model, 0, len(ms))
	for i := range ms {
		models = append(models, &Model{
			ID:           int(ms[i].ID),
			Name:         ms[i].Name,
			Organization: int(ms[i].Organization),
			Created:      *ms[i].Created,
			Updated:      *ms[i].Updated,
		})
	}
	s.httpJSON(w, models, http.StatusOK)
}

func (s *server) ModelsCreateForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization) {
	body := new(ModelsCreateForOrganizationJSONBody)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(body); err != nil {
		if errors.Is(err, (*json.UnmarshalTypeError)(nil)) {
			s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m, err := s.store.Models(organization).Create(&model.Model{Name: body.Name})
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Model{
		ID:           int(m.ID),
		Name:         m.Name,
		Organization: int(m.Organization),
		Created:      *m.Created,
		Updated:      *m.Updated,
	}, http.StatusCreated)
}

func (s *server) ModelsGetForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, model ParameterModel) {
	m, err := s.store.Models(organization).Get(model)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Model{
		ID:           int(m.ID),
		Name:         m.Name,
		Organization: int(m.Organization),
		Created:      *m.Created,
		Updated:      *m.Updated,
	}, http.StatusOK)
}

func (s *server) SchemasListForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization) {
	ss, err := s.store.Schemas(organization).List()
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	schemas := make([]*Schema, 0, len(ss))
	for i := range ss {
		schemas = append(schemas, &Schema{
			ID:           int(ss[i].ID),
			Name:         ss[i].Name,
			Organization: int(ss[i].Organization),
			Input:        ss[i].Input,
			Output:       ss[i].Output,
			Created:      *ss[i].Created,
			Updated:      *ss[i].Updated,
		})
	}
	s.httpJSON(w, schemas, http.StatusOK)
}

func (s *server) SchemasCreateForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization) {
	body := new(SchemasCreateForOrganizationJSONBody)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(body); err != nil {
		if errors.Is(err, (*json.UnmarshalTypeError)(nil)) {
			s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sl := gojsonschema.NewSchemaLoader()
	sl.Validate = true
	if err := sl.AddSchemas(gojsonschema.NewStringLoader(body.Input)); err != nil {
		s.httpError(w, "invalid input schema", http.StatusUnprocessableEntity)
		return
	}

	if err := sl.AddSchemas(gojsonschema.NewStringLoader(body.Output)); err != nil {
		s.httpError(w, "invalid output schema", http.StatusUnprocessableEntity)
		return
	}

	schema, err := s.store.Schemas(organization).Create(&model.Schema{Input: body.Input, Name: body.Name, Output: body.Output})
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Schema{
		ID:           int(schema.ID),
		Name:         schema.Name,
		Organization: int(schema.Organization),
		Input:        schema.Input,
		Output:       schema.Output,
		Created:      *schema.Created,
		Updated:      *schema.Updated,
	}, http.StatusCreated)
}

func (s *server) SchemasGetForOrganization(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, schema ParameterSchema) {
	sc, err := s.store.Schemas(organization).Get(schema)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Schema{
		ID:           int(sc.ID),
		Name:         sc.Name,
		Organization: int(sc.Organization),
		Input:        sc.Input,
		Output:       sc.Output,
		Created:      *sc.Created,
		Updated:      *sc.Updated,
	}, http.StatusOK)
}

func (s *server) VersionsListForModel(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, model ParameterModel) {
	vs, err := s.store.Versions(organization, model).List()
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versions := make([]*Version, 0, len(vs))
	for i := range vs {
		versions = append(versions, &Version{
			ID:           int(vs[i].ID),
			Name:         vs[i].Name,
			Organization: int(vs[i].Organization),
			Model:        int(vs[i].Model),
			Schema:       int(vs[i].Schema),
			Created:      *vs[i].Created,
			Updated:      *vs[i].Updated,
		})
	}
	s.httpJSON(w, versions, http.StatusOK)
}

func (s *server) VersionsCreateForModel(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, modelParam ParameterModel) {
	body := new(VersionsCreateForModelJSONBody)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(body); err != nil {
		if errors.Is(err, (*json.UnmarshalTypeError)(nil)) {
			s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v, err := s.store.Versions(organization, modelParam).Create(&model.Version{Name: body.Name, Schema: int32(body.Schema)})
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Version{
		ID:           int(v.ID),
		Name:         v.Name,
		Organization: int(v.Organization),
		Model:        int(v.Model),
		Schema:       int(v.Schema),
		Created:      *v.Created,
		Updated:      *v.Updated,
	}, http.StatusCreated)
}

func (s *server) VersionsGetForModel(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, model ParameterModel, version ParameterVersion) {
	v, err := s.store.Versions(organization, model).Get(version)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Version{
		ID:           int(v.ID),
		Name:         v.Name,
		Organization: int(v.Organization),
		Model:        int(v.Model),
		Schema:       int(v.Schema),
		Created:      *v.Created,
		Updated:      *v.Updated,
	}, http.StatusOK)
}

func (s *server) ResultsListForVersion(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, model ParameterModel, version ParameterVersion) {
	rs, err := s.store.Results(organization, model, version).List()
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := make([]*Result, 0, len(rs))
	for i := range rs {
		results = append(results, &Result{
			ID:           int(rs[i].ID),
			Organization: int(rs[i].Organization),
			Model:        int(rs[i].Model),
			Version:      int(rs[i].Version),
			Input:        rs[i].Input,
			Output:       rs[i].Output,
			TrueOutput:   rs[i].TrueOutput,
			Time:         rs[i].Time,
			Created:      *rs[i].Created,
			Updated:      *rs[i].Updated,
		})
	}
	s.httpJSON(w, results, http.StatusOK)
}

func (s *server) ResultsCreateForVersion(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, modelParam ParameterModel, version ParameterVersion) {
	v, err := s.store.Versions(organization, modelParam).Get(version)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	schema, err := s.store.Schemas(organization).GetByID(int(v.Schema))
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body := new(ResultsCreateForVersionJSONBody)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(body); err != nil {
		if errors.Is(err, (*json.UnmarshalTypeError)(nil)) {
			s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inputSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema.Input))
	if err != nil {
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	validationResult, err := inputSchema.Validate(gojsonschema.NewStringLoader(body.Input))
	if err != nil {
		s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !validationResult.Valid() {
		s.httpError(w, "input does not match input schema", http.StatusUnprocessableEntity)
		return
	}

	outputSchema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(schema.Output))
	if err != nil {
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	validationResult, err = outputSchema.Validate(gojsonschema.NewStringLoader(body.Output))
	if err != nil {
		s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !validationResult.Valid() {
		s.httpError(w, "output does not match output schema", http.StatusUnprocessableEntity)
		return
	}
	validationResult, err = outputSchema.Validate(gojsonschema.NewStringLoader(body.TrueOutput))
	if err != nil {
		s.httpError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !validationResult.Valid() {
		s.httpError(w, "true output does not match output schema", http.StatusUnprocessableEntity)
		return
	}

	if body.Time == nil {
		t := time.Now()
		body.Time = &t
	}

	result, err := s.store.Results(organization, modelParam, version).Create(&model.Result{Input: body.Input, Output: body.Output, TrueOutput: body.TrueOutput, Time: *body.Time})
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Result{
		ID:           int(result.ID),
		Organization: int(result.Organization),
		Model:        int(result.Model),
		Version:      int(result.Version),
		Input:        result.Input,
		Output:       result.Output,
		TrueOutput:   result.TrueOutput,
		Time:         result.Time,
		Created:      *result.Created,
		Updated:      *result.Updated,
	}, http.StatusCreated)
}

func (s *server) ResultsGetForVersion(w http.ResponseWriter, r *http.Request, organization ParameterOrganization, model ParameterModel, version ParameterVersion, result ParameterResult) {
	v, err := s.store.Versions(organization, model).Get(version)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			s.httpError(w, err.Error(), http.StatusNotFound)
			return
		}
		s.httpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.httpJSON(w, &Version{
		ID:           int(v.ID),
		Name:         v.Name,
		Organization: int(v.Organization),
		Model:        int(v.Model),
		Schema:       int(v.Schema),
		Created:      *v.Created,
		Updated:      *v.Updated,
	}, http.StatusOK)
}
