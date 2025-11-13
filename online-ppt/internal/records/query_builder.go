package records

import (
	"fmt"
	"strings"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

type listQueryBuilder struct {
	filters    ListFilters
	whereItems []string
	args       []any
	sortClause string
}

func normalizeListFilters(filters ListFilters) ListFilters {
	normalized := ListFilters{
		Query:  strings.TrimSpace(filters.Query),
		Tag:    strings.TrimSpace(filters.Tag),
		Sort:   strings.TrimSpace(filters.Sort),
		Limit:  filters.Limit,
		Offset: filters.Offset,
	}

	if normalized.Tag != "" {
		normalized.Tag = strings.ToLower(normalized.Tag)
	}

	switch normalized.Sort {
	case "created_at_asc", "created_at_desc", "name_asc":
		// accepted values
	case "name_desc":
		// allow optional descending sort by name even if not advertised
	default:
		normalized.Sort = "created_at_desc"
	}

	if normalized.Limit <= 0 || normalized.Limit > maxListLimit {
		normalized.Limit = defaultListLimit
	}

	if normalized.Offset < 0 {
		normalized.Offset = 0
	}

	return normalized
}

func newListQueryBuilder(userID int64, filters ListFilters) *listQueryBuilder {
	builder := &listQueryBuilder{
		filters:    filters,
		whereItems: []string{"user_id = ?"},
		args:       []any{userID},
		sortClause: "ORDER BY created_at DESC",
	}

	builder.applyQuery()
	builder.applyTag()
	builder.applySort()

	return builder
}

func (b *listQueryBuilder) selectQuery() (string, []any) {
	var query strings.Builder
	query.WriteString(`SELECT id, user_id, name, title, description, group_name, relative_path, canonical_path, tags, created_at, updated_at FROM ppt_records`)
	query.WriteString(b.whereClause())
	query.WriteRune(' ')
	query.WriteString(b.sortClause)
	query.WriteString(` LIMIT ? OFFSET ?`)

	args := append([]any{}, b.args...)
	args = append(args, b.filters.Limit, b.filters.Offset)
	return query.String(), args
}

func (b *listQueryBuilder) countQuery() (string, []any) {
	var query strings.Builder
	query.WriteString(`SELECT COUNT(*) FROM ppt_records`)
	query.WriteString(b.whereClause())
	return query.String(), append([]any{}, b.args...)
}

func (b *listQueryBuilder) whereClause() string {
	if len(b.whereItems) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.whereItems, " AND ")
}

func (b *listQueryBuilder) applyQuery() {
	if b.filters.Query == "" {
		return
	}
	b.whereItems = append(b.whereItems, "(name LIKE ? OR title LIKE ? OR description LIKE ?)")
	like := fmt.Sprintf("%%%s%%", b.filters.Query)
	b.args = append(b.args, like, like, like)
}

func (b *listQueryBuilder) applyTag() {
	if b.filters.Tag == "" {
		return
	}
	b.whereItems = append(b.whereItems, "JSON_CONTAINS(IFNULL(tags, JSON_ARRAY()), JSON_QUOTE(?))")
	b.args = append(b.args, b.filters.Tag)
}

func (b *listQueryBuilder) applySort() {
	switch b.filters.Sort {
	case "created_at_asc":
		b.sortClause = "ORDER BY created_at ASC"
	case "name_asc":
		b.sortClause = "ORDER BY name ASC"
	case "name_desc":
		b.sortClause = "ORDER BY name DESC"
	default:
		b.sortClause = "ORDER BY created_at DESC"
	}
}
