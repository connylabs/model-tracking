//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Organization = newOrganizationTable("public", "organization", "")

type organizationTable struct {
	postgres.Table

	//Columns
	ID      postgres.ColumnInteger
	Name    postgres.ColumnString
	Created postgres.ColumnTimestamp
	Updated postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type OrganizationTable struct {
	organizationTable

	EXCLUDED organizationTable
}

// AS creates new OrganizationTable with assigned alias
func (a OrganizationTable) AS(alias string) *OrganizationTable {
	return newOrganizationTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new OrganizationTable with assigned schema name
func (a OrganizationTable) FromSchema(schemaName string) *OrganizationTable {
	return newOrganizationTable(schemaName, a.TableName(), a.Alias())
}

func newOrganizationTable(schemaName, tableName, alias string) *OrganizationTable {
	return &OrganizationTable{
		organizationTable: newOrganizationTableImpl(schemaName, tableName, alias),
		EXCLUDED:          newOrganizationTableImpl("", "excluded", ""),
	}
}

func newOrganizationTableImpl(schemaName, tableName, alias string) organizationTable {
	var (
		IDColumn       = postgres.IntegerColumn("id")
		NameColumn     = postgres.StringColumn("name")
		CreatedColumn  = postgres.TimestampColumn("created")
		UpdatedColumn  = postgres.TimestampColumn("updated")
		allColumns     = postgres.ColumnList{IDColumn, NameColumn, CreatedColumn, UpdatedColumn}
		mutableColumns = postgres.ColumnList{NameColumn, CreatedColumn, UpdatedColumn}
	)

	return organizationTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:      IDColumn,
		Name:    NameColumn,
		Created: CreatedColumn,
		Updated: UpdatedColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
