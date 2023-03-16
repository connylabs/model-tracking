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

var Schema = newSchemaTable("public", "schema", "")

type schemaTable struct {
	postgres.Table

	//Columns
	ID           postgres.ColumnInteger
	Name         postgres.ColumnString
	Input        postgres.ColumnString
	Output       postgres.ColumnString
	Organization postgres.ColumnInteger
	Created      postgres.ColumnTimestamp
	Updated      postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type SchemaTable struct {
	schemaTable

	EXCLUDED schemaTable
}

// AS creates new SchemaTable with assigned alias
func (a SchemaTable) AS(alias string) *SchemaTable {
	return newSchemaTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new SchemaTable with assigned schema name
func (a SchemaTable) FromSchema(schemaName string) *SchemaTable {
	return newSchemaTable(schemaName, a.TableName(), a.Alias())
}

func newSchemaTable(schemaName, tableName, alias string) *SchemaTable {
	return &SchemaTable{
		schemaTable: newSchemaTableImpl(schemaName, tableName, alias),
		EXCLUDED:    newSchemaTableImpl("", "excluded", ""),
	}
}

func newSchemaTableImpl(schemaName, tableName, alias string) schemaTable {
	var (
		IDColumn           = postgres.IntegerColumn("id")
		NameColumn         = postgres.StringColumn("name")
		InputColumn        = postgres.StringColumn("input")
		OutputColumn       = postgres.StringColumn("output")
		OrganizationColumn = postgres.IntegerColumn("organization")
		CreatedColumn      = postgres.TimestampColumn("created")
		UpdatedColumn      = postgres.TimestampColumn("updated")
		allColumns         = postgres.ColumnList{IDColumn, NameColumn, InputColumn, OutputColumn, OrganizationColumn, CreatedColumn, UpdatedColumn}
		mutableColumns     = postgres.ColumnList{NameColumn, InputColumn, OutputColumn, OrganizationColumn, CreatedColumn, UpdatedColumn}
	)

	return schemaTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:           IDColumn,
		Name:         NameColumn,
		Input:        InputColumn,
		Output:       OutputColumn,
		Organization: OrganizationColumn,
		Created:      CreatedColumn,
		Updated:      UpdatedColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
