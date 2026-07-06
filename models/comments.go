// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
)

// ColumnComment - ColumnComment type
type ColumnComment struct {
	ColumnName string `json:"columnName"`
	Comment    string `json:"comment"`
}

// Column - Comment type
type Column struct {
	ColumnName string `json:"columnName"`
}

// TableComment - TableComment type
type TableComment struct {
	Comment string `json:"comment"`
}

// BuildSetTableCommentQuery - util
func BuildSetTableCommentQuery(tableName string, comment string) string {
	return fmt.Sprintf(`COMMENT ON TABLE %s IS '%s';`, tableName, comment)
}

// BuildDeleteTableCommentQuery - util
func BuildDeleteTableCommentQuery(tableName string) string {
	return fmt.Sprintf(`COMMENT ON TABLE %s IS NULL;`, tableName)
}

// BuildSetColumnCommentQuery - util
func BuildSetColumnCommentQuery(tableName string, columnName string, comment string) string {
	return fmt.Sprintf(`COMMENT ON COLUMN %s.%s IS '%s';`, tableName, columnName, comment)
}

// BuildDeleteColumnCommentQuery - util
func BuildDeleteColumnCommentQuery(tableName string, columnName string) string {
	return fmt.Sprintf(`COMMENT ON COLUMN %s.%s IS NULL;`, tableName, columnName)
}
