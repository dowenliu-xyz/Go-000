package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/dialect/entsql"
	"github.com/facebook/ent/schema"
	"github.com/facebook/ent/schema/field"
)

// CustomerFollow holds the schema definition for the CustomerFollow entity.
type CustomerFollow struct {
	ent.Schema
}

// Annotations of the CustomerFollow
func (CustomerFollow) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "customer_follow",
		},
	}
}

// Fields of the CustomerFollow.
func (CustomerFollow) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("staffId").StorageKey("staff_id"),
		field.Int64("customerId").StorageKey("customer_id"),
		field.String("content"),
		field.Time("createTime").StorageKey("create_time"),
		field.Int64("deleted"),
		field.String("deleteBy").Optional().StorageKey("deleted_by"),
	}
}

// Edges of the CustomerFollow.
func (CustomerFollow) Edges() []ent.Edge {
	return nil
}
