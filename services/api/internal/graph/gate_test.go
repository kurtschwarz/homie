package graph

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

// TestRootFieldsAreClassified is the test the operation gate's design calls
// for: a root field that is in neither classification set fails the build,
// rather than silently becoming reachable or silently becoming protected.
func TestRootFieldsAreClassified(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	check := func(defName string, def *ast.Definition, public, protected map[string]bool) {
		t.Helper()
		if def == nil {
			return
		}
		for _, field := range def.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			isPublic, isProtected := public[field.Name], protected[field.Name]
			if isPublic && isProtected {
				t.Errorf("%s.%s is classified as both public and protected", defName, field.Name)
			}
			if !isPublic && !isProtected {
				t.Errorf("%s.%s is classified as neither public nor protected", defName, field.Name)
			}
		}
	}

	check("Query", schema.Query, publicQueryFields, protectedQueryFields)
	check("Mutation", schema.Mutation, publicMutationFields, protectedMutationFields)
}
