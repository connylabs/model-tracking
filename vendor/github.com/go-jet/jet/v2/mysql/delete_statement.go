package mysql

import "github.com/go-jet/jet/v2/internal/jet"

// DeleteStatement is interface for MySQL DELETE statement
type DeleteStatement interface {
	Statement

	USING(tables ...ReadableTable) DeleteStatement
	WHERE(expression BoolExpression) DeleteStatement
	ORDER_BY(orderByClauses ...OrderByClause) DeleteStatement
	LIMIT(limit int64) DeleteStatement
}

type deleteStatementImpl struct {
	jet.SerializerStatement

	Delete  jet.ClauseStatementBegin
	Using   jet.ClauseFrom
	Where   jet.ClauseWhere
	OrderBy jet.ClauseOrderBy
	Limit   jet.ClauseLimit
}

func newDeleteStatement(table Table) DeleteStatement {
	newDelete := &deleteStatementImpl{}
	newDelete.SerializerStatement = jet.NewStatementImpl(Dialect, jet.DeleteStatementType, newDelete,
		&newDelete.Delete,
		&newDelete.Using,
		&newDelete.Where,
		&newDelete.OrderBy,
		&newDelete.Limit)

	newDelete.Delete.Name = "DELETE FROM"
	newDelete.Using.Name = "USING"
	newDelete.Delete.Tables = append(newDelete.Delete.Tables, table)
	newDelete.Where.Mandatory = true
	newDelete.Limit.Count = -1

	return newDelete
}

func (d *deleteStatementImpl) USING(tables ...ReadableTable) DeleteStatement {
	d.Using.Tables = readableTablesToSerializerList(tables)
	return d
}

func (d *deleteStatementImpl) WHERE(expression BoolExpression) DeleteStatement {
	d.Where.Condition = expression
	return d
}

func (d *deleteStatementImpl) ORDER_BY(orderByClauses ...OrderByClause) DeleteStatement {
	d.OrderBy.List = orderByClauses
	return d
}

func (d *deleteStatementImpl) LIMIT(limit int64) DeleteStatement {
	d.Limit.Count = limit
	return d
}
