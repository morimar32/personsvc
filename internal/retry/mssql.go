package retry

import (
	mssql "github.com/denisenkom/go-mssqldb"
)

func WithMSSQLSupport() DbRetryOption {
	return func(o *DbRetry) {
		o.evalError = func(err error) bool {
			if mssqlerr, ok := err.(mssql.Error); ok {
				switch mssqlerr.Number {
				case 1205, 1231: // deadlock
					return true
				case -2: // timeout
				default:
					return false
				}
			}
			return false
		}
	}
}
