// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

// IQQuestion - IQQuestion type
type IQQuestion struct {
	Tables   []string `form:"tables" json:"tables" xml:"tables" query:"tables"`
	Question string   `form:"question" json:"question" xml:"question" query:"question"`
}

// IQQuestionDTO - IQQuestionDTO type
type IQQuestionDTO struct {
	Tables    []string `json:"tables"`
	Question  string   `json:"question"`
	SessionID string   `json:"session_id"`
}

// IQQuestionResponseDTO - IQQuestionResponseDTO type
type IQQuestionResponseDTO struct {
	Answer        string   `json:"answer"`
	SQL           string   `json:"sql"`
	SQLComplexity int      `json:"sql_complexity"`
	FeedbackID    string   `json:"feedback_id"`
	SnippetIDs    []string `json:"snippet_ids"`
}

// IQQuery - IQQuery type
type IQQuery struct {
	Tables   []string `form:"tables" json:"tables" xml:"tables" query:"tables"`
	Question string   `form:"question" json:"question" xml:"question" query:"question"`
}

// IQQueryDTO - IQQueryDTO type
type IQQueryDTO struct {
	Tables    []string `json:"tables"`
	Question  string   `json:"question"`
	SessionID string   `json:"session_id"`
}

// IQAutoQueryDTO - IQQueryDTO type
type IQAutoQueryDTO struct {
	Question      string   `json:"question"`
	SessionID     string   `json:"session_id"`
	AllowedTables []string `json:"allowed_tables"`
}

// IQQueryResponseDTO - IQQueryResponseDTO type
type IQQueryResponseDTO struct {
	SQL           string   `json:"sql"`
	SQLComplexity int      `json:"sql_complexity"`
	FeedbackID    string   `json:"feedback_id"`
	SnippetIDs    []string `json:"snippet_ids"`
}

// IQAutoQueryResponseDTO - IQAutoQueryResponseDTO type
type IQAutoQueryResponseDTO struct {
	SQL           string   `json:"sql"`
	SQLComplexity int      `json:"sql_complexity"`
	FeedbackID    string   `json:"feedback_id"`
	Tables        []string `json:"tables"`
	SnippetIDs    []string `json:"snippet_ids"`
}

// IQAnswer - IQAnswer type
type IQAnswer struct {
	Tables   []string `form:"tables" json:"tables" xml:"tables" query:"tables"`
	Query    string   `form:"query" json:"query" xml:"query" query:"query"`
	Question string   `form:"question" json:"question" xml:"question" query:"question"`
}

// IQAnswerDTO - IQAnswerDTO type
type IQAnswerDTO struct {
	Query     string   `json:"query"`
	Tables    []string `json:"tables"`
	Question  string   `json:"question"`
	SessionID string   `json:"session_id"`
}

// IQAnswerResponseDTO - IQAnswerResponseDTO type
type IQAnswerResponseDTO struct {
	Answer        string `json:"answer"`
	SQL           string `json:"sql"`
	SQLResults    string `json:"sql_results"`
	SQLComplexity int    `json:"sql_complexity"`
	FeedbackID    string `json:"feedback_id"`
}

// IQErrorResponseDTO - IQErrorResponseDTO type
type IQErrorResponseDTO struct {
	Error string `json:"error"`
	SQL   string `json:"sql"`
}

// IQFeedbackDTO - IQFeedbackDTO type
type IQFeedbackDTO struct {
	FeedbackID string  `json:"feedback_id"`
	Score      float32 `json:"score"`
	Comment    string  `json:"comment"`
}
