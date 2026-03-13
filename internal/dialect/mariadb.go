package dialect

// MariaDBDialect implements Dialect for MariaDB by wrapping MySQLDialect.
type MariaDBDialect struct {
	MySQLDialect
}

func (d *MariaDBDialect) Name() string {
	return "mariadb"
}
