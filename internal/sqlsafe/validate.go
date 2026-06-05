// Package sqlsafe provides SQL read-only validation shared across bi and genkit.
// Single source of truth for the rule: only SELECT, single-statement, no DDL/DML.
package sqlsafe

import (
	"fmt"
	"strings"
)

/*
	TODO: This validate function is used by DomusExecutor struct to validate SQL queries.
	Move it to DomusExecutor struct when you have time.

	Any other method in Domus already validate SQL (ex: datastores, ORM querys, ...),
	so if we move this to DomusExecutor struct, we can prevent raw query execution sended from .Exec method.
	The point is: any user could use .Exec method directly without any validation,
	if we didn't validate SQL in DomusExecutor struct, it will be a security risk, because this sqlsafe package
	can be bypassed easily by someone how know this internal package.
*/

// forbiddenKeywords are DML/DDL verbs that must not appear in read-only SQL.
// Each entry is surrounded by spaces in the check so that column names like
// "last_updated" are not incorrectly rejected.
var forbiddenKeywords = []string{
	" insert ", " update ", " delete ",
	" drop ", " alter ", " truncate ", " create ",
}

// Validate returns an error if sql is not a safe, read-only single-statement query.
// It is intentionally strict: any doubt → reject.
func Validate(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return fmt.Errorf("sql is required")
	}

	lower := strings.ToLower(trimmed)
	for _, kw := range forbiddenKeywords {
		if strings.Contains(" "+lower+" ", kw) {
			return fmt.Errorf("sql must be read-only (rejected keyword: %q)", strings.TrimSpace(kw))
		}
	}

	// Reject multi-statement: semicolons anywhere except as the final character.
	if strings.Count(trimmed, ";") > 1 ||
		(strings.Contains(trimmed, ";") && !strings.HasSuffix(trimmed, ";")) {
		return fmt.Errorf("sql must be a single statement")
	}

	return nil
}
