package knowledge

import (
	"time"
)

type EntityType string

const (
	EntityTypePerson  EntityType = "person"
	EntityTypeTopic   EntityType = "topic"
	EntityTypeConcept EntityType = "concept"
	EntityTypeEvent   EntityType = "event"
)

type Entity struct {
	ID          int64
	SessionID   string
	Type        EntityType
	Name        string
	Description string
	Keywords    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RelationType string

const (
	RelationRelatedTo   RelationType = "related_to"
	RelationDependsOn   RelationType = "depends_on"
	RelationContradicts RelationType = "contradicts"
	RelationPartOf      RelationType = "part_of"
)

type Relation struct {
	ID           int64
	SessionID    string
	FromEntityID int64
	ToEntityID   int64
	Type         RelationType
	Context      string
	CreatedAt    time.Time
}

type KnowledgePage struct {
	ID             int64
	SessionID      string
	EntityID       int64
	Title          string
	Content        string
	SourceMessages string
	Version        int
	IsDraft        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type LogEntry struct {
	ID        int64
	SessionID string
	Action    string
	Detail    string
	CreatedAt time.Time
}

const (
	ActionIngest = "ingest"
	ActionQuery  = "query"
	ActionLint   = "lint"
	ActionUpdate = "update"
)
