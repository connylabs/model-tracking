package store

import (
	"context"

	"github.com/connylabs/model-tracking/store/model-tracking/public/model"
)

// ModelTracking is a store that can return all other kinds of stores.
type ModelTracking interface {
	// Organizations returns a store for interacting with organizations.
	Organizations() Organizations
	// Models returns a store for interacting with models.
	Models(organization string) Models
	// Schemas returns a store for interacting with schemas.
	Schemas(organization string) Schemas
	// Versions returns a store for interacting with versions.
	Versions(organization, model string) Versions
	// Results returns a store for interacting with results.
	Results(organization, model, version string) Results
}

// Organizations is a store that allows interacting with organizations.
type Organizations interface {
	// Create creates a new organization in the store.
	Create(context.Context, *model.Organization) (*model.Organization, error)
}

// Models is a store that allows interacting with models.
type Models interface {
	// Create creates a new model for the organization in the store.
	Create(context.Context, *model.Model) (*model.Model, error)
	// Update updates a model for the organization in the store.
	Update(context.Context, *model.Model) (*model.Model, error)
	// Get gets a model for the organization in the store.
	Get(ctx context.Context, name string) (*model.Model, error)
	// List gets all models for the organization in the store.
	List(context.Context) ([]*model.Model, error)
}

// Schemas is a store that allows interacting with schemas.
type Schemas interface {
	// Create creates a new schemas for the organization in the store.
	Create(context.Context, *model.Schema) (*model.Schema, error)
	// Get gets a schema for the organization in the store.
	Get(ctx context.Context, name string) (*model.Schema, error)
	// GetByID gets a schema for the organization in the store.
	GetByID(ctx context.Context, id int) (*model.Schema, error)
	// List gets all schemas for the organization in the store.
	List(context.Context) ([]*model.Schema, error)
}

// Versions is a store that allows interacting with versions.
type Versions interface {
	// Create creates a new versions for the model in the store.
	Create(context.Context, *model.Version) (*model.Version, error)
	// Get gets a version for the model in the store.
	Get(ctx context.Context, name string) (*model.Version, error)
	// GetOrCreate gets a version for the model in the store and if it does not exist,
	// it tries to create it using the default schema of the model.
	GetOrCreate(ctx context.Context, name string) (*model.Version, error)
	// List gets all versions for the model in the store.
	List(context.Context) ([]*model.Version, error)
}

// Results is a store that allows interacting with results.
type Results interface {
	// Create creates a new result for a version of the model in the store.
	Create(context.Context, *model.Result) (*model.Result, error)
	// Get gets a result for a version of the model in the store.
	Get(ctx context.Context, id int) (*model.Result, error)
	// List gets all results for a version the model in the store.
	List(context.Context) ([]*model.Result, error)
}
