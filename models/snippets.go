// SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package models

// Snippet represents a single snippet.
type Snippet struct {
	SnippetID string `json:"snippet_id"`
	Snippet   string `json:"snippet"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SnippetInsertRequest represents the request body for inserting a snippet.
type SnippetInsertRequest struct {
	Snippets  []string `json:"snippets"`
	SessionID string   `json:"session_id"`
}

// SnippetInsertResponse represents the response body for inserting a snippet.
type SnippetInsertResponse struct {
	SnippetIDs []string `json:"snippet_ids"`
}

// SnippetUpdateRequest represents the request body for updating a snippet.
type SnippetUpdateRequest struct {
	Snippet   string `json:"snippet"`
	SnippetID string `json:"snippet_id"`
	SessionID string `json:"session_id"`
}

// SnippetUpdateResponse represents the response body for updating a snippet.
type SnippetUpdateResponse struct {
	SnippetID string `json:"snippet_id"`
}

// SnippetGetRequest represents the request body for getting a snippet.
type SnippetGetRequest struct {
	SnippetID string `json:"snippet_id"`
	SessionID string `json:"session_id"`
}

// SnippetGetResponse represents the response body for getting a snippet.
type SnippetGetResponse struct {
	Snippet   string `json:"snippet"`
	SnippetID string `json:"snippet_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// SnippetListRequest represents the request body for listing snippets.
type SnippetListRequest struct {
	SessionID string `json:"session_id"`
}

// SnippetListResponse represents the response body for listing snippets.
type SnippetListResponse struct {
	Snippets []Snippet `json:"snippets"`
}

// SnippetDeleteRequest represents the request body for deleting snippets.
type SnippetDeleteRequest struct {
	SnippetIDs []string `json:"snippet_ids"`
	SessionID  string   `json:"session_id"`
}

// SnippetDeleteResponse represents the response body for deleting snippets.
type SnippetDeleteResponse struct {
	Deleted bool `json:"deleted"`
}
