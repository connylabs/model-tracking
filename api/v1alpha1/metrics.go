// This code in auto generated. DO NOT EDIT.

package v1alpha1

import (
	"net/http"

	"github.com/metalmatze/signal/server/signalhttp"
	"github.com/prometheus/client_golang/prometheus"
)

type InstrumentedServerInterface struct {
	ServerInterface
	signalhttp.HandlerInstrumenter
}

func NewInstrumentedServerInterface(impl ServerInterface, r prometheus.Registerer) *InstrumentedServerInterface {
	i := signalhttp.NewHandlerInstrumenter(r, []string{"handler"})
	return &InstrumentedServerInterface{impl, i}
}

func (i *InstrumentedServerInterface) ModelsCreateForOrganization(w http.ResponseWriter, r *http.Request, _c2 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ModelsCreateForOrganization(w, r, _c2)
	}
	i.NewHandler(prometheus.Labels{"handler": "ModelsCreateForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ModelsGetForOrganization(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ModelsGetForOrganization(w, r, _c2, _c3)
	}
	i.NewHandler(prometheus.Labels{"handler": "ModelsGetForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ModelsListForOrganization(w http.ResponseWriter, r *http.Request, _c2 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ModelsListForOrganization(w, r, _c2)
	}
	i.NewHandler(prometheus.Labels{"handler": "ModelsListForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ModelsUpdateForOrganization(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ModelsUpdateForOrganization(w, r, _c2, _c3)
	}
	i.NewHandler(prometheus.Labels{"handler": "ModelsUpdateForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) OrganizationsCreate(w http.ResponseWriter, r *http.Request) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.OrganizationsCreate(w, r)
	}
	i.NewHandler(prometheus.Labels{"handler": "OrganizationsCreate"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ResultsCreateForVersion(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string, _c4 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ResultsCreateForVersion(w, r, _c2, _c3, _c4)
	}
	i.NewHandler(prometheus.Labels{"handler": "ResultsCreateForVersion"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ResultsGetForVersion(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string, _c4 string, _c5 int) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ResultsGetForVersion(w, r, _c2, _c3, _c4, _c5)
	}
	i.NewHandler(prometheus.Labels{"handler": "ResultsGetForVersion"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) ResultsListForVersion(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string, _c4 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.ResultsListForVersion(w, r, _c2, _c3, _c4)
	}
	i.NewHandler(prometheus.Labels{"handler": "ResultsListForVersion"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) SchemasCreateForOrganization(w http.ResponseWriter, r *http.Request, _c2 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.SchemasCreateForOrganization(w, r, _c2)
	}
	i.NewHandler(prometheus.Labels{"handler": "SchemasCreateForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) SchemasGetForOrganization(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.SchemasGetForOrganization(w, r, _c2, _c3)
	}
	i.NewHandler(prometheus.Labels{"handler": "SchemasGetForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) SchemasListForOrganization(w http.ResponseWriter, r *http.Request, _c2 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.SchemasListForOrganization(w, r, _c2)
	}
	i.NewHandler(prometheus.Labels{"handler": "SchemasListForOrganization"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) VersionsCreateForModel(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.VersionsCreateForModel(w, r, _c2, _c3)
	}
	i.NewHandler(prometheus.Labels{"handler": "VersionsCreateForModel"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) VersionsGetForModel(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string, _c4 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.VersionsGetForModel(w, r, _c2, _c3, _c4)
	}
	i.NewHandler(prometheus.Labels{"handler": "VersionsGetForModel"}, http.HandlerFunc(handler))(w, r)
}

func (i *InstrumentedServerInterface) VersionsListForModel(w http.ResponseWriter, r *http.Request, _c2 string, _c3 string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		i.ServerInterface.VersionsListForModel(w, r, _c2, _c3)
	}
	i.NewHandler(prometheus.Labels{"handler": "VersionsListForModel"}, http.HandlerFunc(handler))(w, r)
}
