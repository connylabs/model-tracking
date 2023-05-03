package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/efficientgo/core/testutil"
	"github.com/efficientgo/e2e"

	"github.com/connylabs/model-tracking/api/v1alpha1"
)

const (
	httpPortName        = "http"
	metricsPortName     = "metrics"
	postgresUser        = "user"
	postgresPassword    = "pass"
	postgresURLTemplate = "postgres://%s:%s@%s/%s?sslmode=disable"
)

func newPostgres(env e2e.Environment, name, db string) e2e.Runnable {
	return env.Runnable(name).WithPorts(map[string]int{httpPortName: 5432}).Init(
		e2e.StartOptions{
			Image: "postgres:15.2",
			EnvVars: map[string]string{
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPassword,
				"POSTGRES_DB":       db,
			},
			Readiness: e2e.NewCmdReadinessProbe(e2e.NewCommand("psql", "-U", postgresUser, "-d", db, "-c", "SELECT 1;")),
		},
	)
}

func databaseURL(endpoint, database string) string {
	return fmt.Sprintf(postgresURLTemplate, postgresUser, postgresPassword, endpoint, database)
}

func migratePostgres(ctx context.Context, dbURL string) error {
	migrate := exec.CommandContext(ctx, "make", "migrate")
	migrate.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbURL),
	)
	out, err := migrate.CombinedOutput()
	if err != nil {
		return errors.Join(fmt.Errorf("failed to execute migration; context:\n%s", string(out)), err)
	}
	return nil
}

func newModelTracking(env e2e.Environment, name, dbURL string) e2e.Runnable {
	return env.Runnable(name).WithPorts(map[string]int{
		httpPortName:    8080,
		metricsPortName: 9090,
	}).Init(
		e2e.StartOptions{
			Image:     "ghcr.io/connylabs/model-tracking",
			Readiness: e2e.NewHTTPReadinessProbe(metricsPortName, "/ready", 200, 200),
			Command: e2e.Command{
				Args: []string{
					"--database",
					dbURL,
				},
			},
		},
	)
}

func newSetup(t *testing.T, name string) e2e.Runnable {
	e, err := e2e.New(e2e.WithName("e2e-" + name))
	testutil.Ok(t, err)
	t.Cleanup(e.Close)

	db := "model-tracking"
	postgresContainer := newPostgres(e, "postgres", db)
	testutil.Ok(t, e2e.StartAndWaitReady(postgresContainer))
	testutil.Ok(t, migratePostgres(context.Background(), databaseURL(postgresContainer.Endpoint(httpPortName), db)))

	modelTrackingContainer := newModelTracking(e, "model-tracking", databaseURL(postgresContainer.InternalEndpoint(httpPortName), db))
	testutil.Ok(t, e2e.StartAndWaitReady(modelTrackingContainer))

	return modelTrackingContainer
}

func TestStartup(t *testing.T) {
	t.Parallel()

	newSetup(t, "startup")
}

func mustBody(t *testing.T, body io.Reader) string {
	b, err := io.ReadAll(body)
	testutil.Ok(t, err)
	return string(b)
}

func mustRequest(r *http.Request, err error) *http.Request {
	if err != nil {
		panic(err)
	}
	return r
}

func registerOrganizationModelSchemaVersion(t *testing.T, c *v1alpha1.ClientWithResponses, o, m, s, v string, input, output json.RawMessage) {
	r, err := c.OrganizationsCreate(context.Background(), v1alpha1.OrganizationsCreateJSONRequestBody{Name: o})

	testutil.Ok(t, err)
	testutil.Equals(t, 201, r.StatusCode, "status should be 201, got %d with body:\n %s", r.StatusCode, mustBody(t, r.Body))

	r, err = c.ModelsCreateForOrganization(context.Background(), o, v1alpha1.ModelsCreateForOrganizationJSONRequestBody{Name: m})
	testutil.Ok(t, err)
	testutil.Equals(t, 201, r.StatusCode, "status should be 201, got %d with body:\n %s", r.StatusCode, mustBody(t, r.Body))

	schema, err := c.SchemasCreateForOrganizationWithResponse(context.Background(), o, v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: s, Input: input, Output: output})
	testutil.Ok(t, err)
	testutil.Equals(t, 201, r.StatusCode, "status should be 201, got %d with body:\n %s", r.StatusCode, mustBody(t, r.Body))

	r, err = c.VersionsCreateForModel(context.Background(), o, m, v1alpha1.VersionsCreateForModelJSONRequestBody{Name: v, Schema: schema.JSON201.ID})
	testutil.Ok(t, err)
	testutil.Equals(t, 201, r.StatusCode, "status should be 201, got %d with body:\n %s", r.StatusCode, mustBody(t, r.Body))
}

func TestE2E(t *testing.T) {
	t.Parallel()

	modelTrackingContainer := newSetup(t, "createresult")

	server := "http://" + modelTrackingContainer.Endpoint(httpPortName) + "/api/v1alpha1/"
	c, err := v1alpha1.NewClientWithResponses(server)
	testutil.Ok(t, err)

	registerOrganizationModelSchemaVersion(t, c, "foo", "bar", "baz", "qux", inputSchema, outputSchema)

	type request struct {
		request *http.Request
		status  int
	}

	for _, tc := range []struct {
		name     string
		requests []request
	}{
		{
			name: "create result",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "bar", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: output, TrueOutput: output})),
					status:  201,
				},
			},
		},
		{
			name: "invalid result",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "bar", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: []byte(`"cat"`), Output: output, TrueOutput: output})),
					status:  422,
				},
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "bar", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: []byte(`"cat"`), TrueOutput: output})),
					status:  422,
				},
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "bar", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: output, TrueOutput: []byte(`"cat"`)})),
					status:  422,
				},
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "nonexistent-organization", "bar", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: output, TrueOutput: output})),
					status:  404,
				},
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "nonexistent-model", "qux", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: output, TrueOutput: output})),
					status:  404,
				},
				{
					request: mustRequest(v1alpha1.NewResultsCreateForVersionRequest(server, "foo", "bar", "nonexistent-version", v1alpha1.ResultsCreateForVersionJSONRequestBody{Input: input, Output: output, TrueOutput: output})),
					status:  404,
				},
			},
		},
		{
			name: "create version",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewVersionsCreateForModelRequest(server, "foo", "bar", v1alpha1.VersionsCreateForModelJSONRequestBody{Name: "duplicate", Schema: 1})),
					status:  201,
				},
			},
		},
		{
			name: "invalid version",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewVersionsCreateForModelRequest(server, "foo", "bar", v1alpha1.VersionsCreateForModelJSONRequestBody{Name: "qux", Schema: 1})),
					status:  500,
				},
				{
					request: mustRequest(v1alpha1.NewVersionsCreateForModelRequest(server, "foo", "bar", v1alpha1.VersionsCreateForModelJSONRequestBody{Name: "nonexistent-schema", Schema: 311})),
					status:  500,
				},
				{
					request: mustRequest(v1alpha1.NewVersionsCreateForModelRequest(server, "nonexistent-organization", "bar", v1alpha1.VersionsCreateForModelJSONRequestBody{Name: "ok", Schema: 1})),
					status:  404,
				},
				{
					request: mustRequest(v1alpha1.NewVersionsCreateForModelRequest(server, "foo", "nonexistent-model", v1alpha1.VersionsCreateForModelJSONRequestBody{Name: "ok", Schema: 1})),
					status:  404,
				},
			},
		},
		{
			name: "create schema",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewSchemasCreateForOrganizationRequest(server, "foo", v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: "yup", Input: inputSchema, Output: outputSchema})),
					status:  201,
				},
			},
		},
		{
			name: "invalid schema",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewSchemasCreateForOrganizationRequest(server, "foo", v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: "baz", Input: inputSchema, Output: outputSchema})),
					status:  500,
				},
				{
					request: mustRequest(v1alpha1.NewSchemasCreateForOrganizationRequest(server, "foo", v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: "invalidinput", Input: []byte("1"), Output: outputSchema})),
					status:  422,
				},
				{
					request: mustRequest(v1alpha1.NewSchemasCreateForOrganizationRequest(server, "foo", v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: "invalidoutput", Input: inputSchema, Output: []byte("1")})),
					status:  422,
				},
				{
					request: mustRequest(v1alpha1.NewSchemasCreateForOrganizationRequest(server, "nonexistent-organization", v1alpha1.SchemasCreateForOrganizationJSONRequestBody{Name: "ok", Input: inputSchema, Output: outputSchema})),
					status:  404,
				},
			},
		},
		{
			name: "create model",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewModelsCreateForOrganizationRequest(server, "foo", v1alpha1.ModelsCreateForOrganizationJSONRequestBody{Name: "yup"})),
					status:  201,
				},
			},
		},
		{
			name: "invalid model",
			requests: []request{
				{
					request: mustRequest(v1alpha1.NewModelsCreateForOrganizationRequest(server, "foo", v1alpha1.ModelsCreateForOrganizationJSONRequestBody{Name: "bar"})),
					status:  500,
				},
				{
					request: mustRequest(v1alpha1.NewModelsCreateForOrganizationRequest(server, "nonexistent-organization", v1alpha1.ModelsCreateForOrganizationJSONRequestBody{Name: "yup"})),
					status:  404,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, r := range tc.requests {
				res, err := c.ClientInterface.(*v1alpha1.Client).Client.Do(r.request)
				testutil.Ok(t, err)
				testutil.Equals(t, r.status, res.StatusCode, "status should be %d, got %d with body:\n %s", r.status, res.StatusCode, mustBody(t, res.Body))
			}
		})
	}
}

var inputSchema = []byte(`{
  "type": "object",
  "properties": {
    "text": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  },
  "required": [
    "text"
  ],
  "unevaluatedProperties": false
}`)

var input = []byte(`{
  "text": [
    "a",
    "b",
    "c"
  ]
}`)

var outputSchema = []byte(`{
  "type": "object",
  "properties": {
    "predictions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "label": {
            "type": "string"
          },
          "score": {
            "type": "number",
            "maximum": 1,
            "minimum": 0
          }
        }
      }
    }
  },
  "required": [
    "predictions"
  ],
  "unevaluatedProperties": false
}`)

var output = []byte(`{
  "predictions": [
    {
	"label": "cat",
	"score": 0.5
    },
    {
	"label": "kitten",
	"score": 0.25
    }
  ]
}`)
