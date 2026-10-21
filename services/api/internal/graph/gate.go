package graph

import (
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// publicQueryFields and publicMutationFields are reachable with no
// credential at all: a browser has to be able to ask whether it's logged
// in, whether setup is required, and log in before it can prove itself.
//
// Every other root field is protected, and the explicit protected* sets
// below exist only so TestRootFieldsAreClassified can catch a field that
// was added to the schema but never classified either way: the default at
// runtime is already deny, but a forgotten classification should fail the
// build, not pass silently as newly protected.
var publicQueryFields = map[string]bool{
	"session": true,
	"health":  true,
}

var protectedQueryFields = map[string]bool{
	"timer":               true,
	"photos":              true,
	"photoShowing":        true,
	"photoShowingHistory": true,
	"events":              true,
	"people":              true,
	"tagMappings":         true,
	"calendarTags":        true,
	"weather":             true,
	"sessions":            true,
}

var publicMutationFields = map[string]bool{
	"setup": true,
	"login": true,
}

var protectedMutationFields = map[string]bool{
	"startTimer":        true,
	"addTimerMinute":    true,
	"pauseTimer":        true,
	"resumeTimer":       true,
	"cancelTimer":       true,
	"uploadPhoto":       true,
	"trackPhotoShowing": true,
	"createEvent":       true,
	"savePerson":        true,
	"deletePerson":      true,
	"saveTagMapping":    true,
	"deleteTagMapping":  true,
	"refreshEvents":     true,
	"logout":            true,
	"revokeSession":     true,
}

// IsPublicOperation reports whether every root field the operation selects
// is public, meaning it needs no credential at all. A fragment at the root
// is treated conservatively as non-public: nothing selects one today.
func IsPublicOperation(op *ast.OperationDefinition) bool {
	fields := publicQueryFields
	if op.Operation == ast.Mutation {
		fields = publicMutationFields
	}
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			return false
		}
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		if !fields[field.Name] {
			return false
		}
	}
	return true
}
