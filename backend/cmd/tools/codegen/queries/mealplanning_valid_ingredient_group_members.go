package main

import (
	"github.com/cristalhq/builq"
)

const (
	validIngredientGroupMembersTableName            = "valid_ingredient_group_members"
	belongsToGroupColumn                            = "belongs_to_group"
	validIngredientGroupMemberValidIngredientColumn = "valid_ingredient"
)

var validIngredientGroupMembersColumns = []string{
	idColumn,
	belongsToGroupColumn,
	validIngredientGroupMemberValidIngredientColumn,
	createdAtColumn,
	archivedAtColumn,
}

func buildValidIngredientGroupMembersQueries(database string) []*Query {
	switch database {
	case postgres:

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "",
					Type: "",
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(``)),
			},
		}
	default:
		return nil
	}
}
