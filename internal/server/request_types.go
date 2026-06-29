package server

// Request body types for POST/PUT/PATCH endpoints. Kept separate from DTOs so the
// incoming shape can evolve independently of the outgoing shape. Pointer fields on
// the update requests mean "only change this if present" (partial updates).

type createProjectRequest struct {
	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type updateProjectRequest struct {
	Name        *string   `json:"name"`
	Summary     *string   `json:"summary"`
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
}

type promoteRequest struct {
	Mode string `json:"mode"`
}

type setModeRequest struct {
	Mode string `json:"mode"`
}

type archiveRequest struct {
	Reason string `json:"reason"`
}

type reviveRequest struct {
	Mode string `json:"mode"`
}

type upsertDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type createTaskRequest struct {
	ColumnID    string   `json:"column_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Labels      []string `json:"labels"`
	DocumentID  *string  `json:"document_id"`
}

type updateTaskRequest struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Priority    *string   `json:"priority"`
	Labels      *[]string `json:"labels"`
}

type moveTaskRequest struct {
	ColumnID string `json:"column_id"`
	Position int    `json:"position"`
}

type linkTaskDocumentRequest struct {
	DocumentID *string `json:"document_id"`
}
