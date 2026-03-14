package dialect

import (
	"fmt"
	"strings"
)

func (d *MSSQLDialect) CreateTableDDL(tableName, schema string, cols []ColumnDef) string {
	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), fullTableName)
	}

	var sb strings.Builder

	 bareTableName := strings.ToLower(tableName)
	effectiveSchema := strings.ToLower(schema)
	if effectiveSchema == "" {
		effectiveSchema = "dbo"
	}
	sb.WriteString(fmt.Sprintf("IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s')\n", effectiveSchema, bareTableName))
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", fullTableName))

	for i, col := range cols {
		prec := 0
		if col.Precision.Valid {
			prec = int(col.Precision.Int64)
		}
		s := 0
		if col.Scale.Valid {
			s = int(col.Scale.Int64)
		}

		myType := d.MapOracleType(col.Type, prec, s)
		sb.WriteString(fmt.Sprintf("    %s %s", d.QuoteIdentifier(strings.ToLower(col.Name)), myType))

		if col.DefaultValue.Valid && col.DefaultValue.String != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
		}

		if col.Nullable == "N" {
			sb.WriteString(" NOT NULL")
		}
		if i < len(cols)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(");\n")
	return sb.String()
}

func (d *MSSQLDialect) CreateConstraintDDL(constraint ConstraintMetadata, schema string) string {
	table := strings.ToLower(constraint.TableName)
	if schema != "" {
		table = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(table))
	} else {
		table = d.QuoteIdentifier(table)
	}

	name := d.QuoteIdentifier(strings.ToLower(constraint.Name))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s ", table, name))

	if constraint.Type == "R" {
		refTable := strings.ToLower(constraint.RefTableName)
		if schema != "" {
			refTable = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(refTable))
		} else {
			refTable = d.QuoteIdentifier(refTable)
		}

		localCols := make([]string, len(constraint.Columns))
		for i, c := range constraint.Columns {
			localCols[i] = d.QuoteIdentifier(strings.ToLower(c))
		}

		refCols := make([]string, len(constraint.RefColumns))
		for i, c := range constraint.RefColumns {
			refCols[i] = d.QuoteIdentifier(strings.ToLower(c))
		}

		sb.WriteString(fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
			strings.Join(localCols, ", "), refTable, strings.Join(refCols, ", ")))

		if constraint.DeleteRule != "" {
			sb.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.DeleteRule))
		}
	} else if constraint.Type == "C" {
		sb.WriteString(fmt.Sprintf("CHECK (%s)", constraint.SearchCondition))
	}
	sb.WriteString(";\n")
	return sb.String()
}

func (d *MSSQLDialect) CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) {
	name := strings.ToLower(seq.Name)
	if schema != "" {
		name = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(name))
	} else {
		name = d.QuoteIdentifier(name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE SEQUENCE %s\n", name))
	sb.WriteString(fmt.Sprintf("    START WITH %d\n", seq.LastNumber))
	sb.WriteString(fmt.Sprintf("    INCREMENT BY %d\n", seq.IncrementBy))
	sb.WriteString(fmt.Sprintf("    MINVALUE %d\n", seq.MinValue))

	if strings.TrimSpace(seq.MaxValue) != oracleDefaultMaxValue {
		sb.WriteString(fmt.Sprintf("    MAXVALUE %s\n", strings.TrimSpace(seq.MaxValue)))
	}

	if seq.CycleFlag == "Y" {
		sb.WriteString("    CYCLE\n")
	} else {
		sb.WriteString("    NO CYCLE\n")
	}

	sb.WriteString(";\n")
	return sb.String(), true
}

func (d *MSSQLDialect) CreateIndexDDL(idx IndexMetadata, tableName, schema string) string {
	table := strings.ToLower(tableName)
	if schema != "" {
		table = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(table))
	} else {
		table = d.QuoteIdentifier(table)
	}

	colExprs := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		expr := d.QuoteIdentifier(strings.ToLower(col.Name))
		if strings.ToUpper(col.Descend) == "DESC" {
			expr += " DESC"
		}
		colExprs[i] = expr
	}
	colList := strings.Join(colExprs, ", ")

	if idx.IsPK {
		return fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);\n", table, colList)
	}

	indexName := strings.ToLower(idx.Name)
	quotedIndexName := d.QuoteIdentifier(indexName)

	uniqueStr := ""
	if idx.Uniqueness == "UNIQUE" {
		uniqueStr = "UNIQUE "
	}

	bareTableName := strings.ToLower(tableName)
	effectiveSchema := strings.ToLower(schema)
	if effectiveSchema == "" {
		effectiveSchema = "dbo"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"IF NOT EXISTS (\n"+
			"    SELECT 1 FROM sys.indexes i\n"+
			"    JOIN sys.objects o ON i.object_id = o.object_id\n"+
			"    WHERE i.name = '%s'\n"+
			"      AND o.name = '%s'\n"+
			"      AND SCHEMA_NAME(o.schema_id) = '%s'\n"+
			")\n",
		indexName, bareTableName, effectiveSchema,
	))
	sb.WriteString(fmt.Sprintf("    CREATE %sINDEX %s ON %s (%s);\n", uniqueStr, quotedIndexName, table, colList))

	return sb.String()
}
