package graph

import (
	"strings"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

// The directives the schema uses to classify root fields. @public is the
// only one that grants anything: @protected exists so that a field added
// without a decision fails TestRootFieldsAreClassified rather than
// silently inheriting the deny default.
const (
	publicDirective    = "public"
	protectedDirective = "protected"
	kindsArgument      = "kinds"
	roleArgument       = "role"
)

// The AuthKind and AuthRole values the gate knows how to enforce, taken
// from the generated enums so that renaming a value in the schema breaks
// the build here rather than silently matching nothing. Every value must
// be handled: TestEveryAuthKindIsEnforced and TestEveryAuthRoleIsEnforced
// cover adding one, which still compiles.
var (
	kindUser    = string(model.AuthKindUser)
	kindDisplay = string(model.AuthKindDisplay)
	roleAdmin   = string(model.AuthRoleAdmin)
)

// IsPublicOperation reports whether every root field the operation selects
// is marked @public, meaning it needs no credential at all.
//
// The classification is read off the schema rather than a list kept here,
// so the decision sits next to the field it applies to. Everything else
// denies: an unannotated field, a field whose definition the validator
// could not resolve, and a fragment at the root (nothing selects one
// today) are all treated as protected.
func IsPublicOperation(op *ast.OperationDefinition) bool {
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			return false
		}
		if strings.HasPrefix(field.Name, "__") {
			continue
		}
		if field.Definition == nil {
			return false
		}
		if field.Definition.Directives.ForName(publicDirective) == nil {
			return false
		}
	}
	return true
}

// SelectsDisplayOnlyField reports whether the operation touches a root field
// whose @protected kinds exclude USER, which a listener that does not serve
// the display must refuse outright - no credential makes it readable from
// the network.
func SelectsDisplayOnlyField(op *ast.OperationDefinition) bool {
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		kinds := protectedKinds(field.Definition)
		if len(kinds) == 0 {
			continue
		}
		if !containsString(kinds, kindUser) {
			return true
		}
	}
	return false
}

// OperationAllowsKind reports whether every @protected root field the
// operation selects accepts the given kind. A field with no @protected
// directive (public, or unresolved) is not a reason to deny here - that is
// IsPublicOperation's question.
func OperationAllowsKind(op *ast.OperationDefinition, kind string) bool {
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		kinds := protectedKinds(field.Definition)
		if len(kinds) == 0 {
			continue
		}
		if !containsString(kinds, kind) {
			return false
		}
	}
	return true
}

// OperationAllowsRole reports whether every @protected root field the
// operation selects, among those that accept a USER caller at all, is
// satisfied by a USER holding role. A field whose kinds exclude USER is not
// this function's concern - OperationAllowsKind(op, kindUser) already
// denies the operation for a USER caller before role is considered.
func OperationAllowsRole(op *ast.OperationDefinition, role string) bool {
	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}
		kinds := protectedKinds(field.Definition)
		if len(kinds) == 0 || !containsString(kinds, kindUser) {
			continue
		}
		if protectedRole(field.Definition) != role {
			return false
		}
	}
	return true
}

// protectedKinds returns the AuthKind values a field's @protected accepts,
// or nil if the field is not protected at all. An omitted argument means
// [USER], the schema default: gqlparser leaves directive arguments as
// written, so the default is applied here rather than read back out of the
// AST.
func protectedKinds(field *ast.FieldDefinition) []string {
	if field == nil {
		return nil
	}
	directive := field.Directives.ForName(protectedDirective)
	if directive == nil {
		return nil
	}
	argument := directive.Arguments.ForName(kindsArgument)
	if argument == nil || argument.Value == nil {
		return []string{kindUser}
	}
	kinds := make([]string, 0, len(argument.Value.Children))
	for _, child := range argument.Value.Children {
		kinds = append(kinds, child.Value.Raw)
	}
	return kinds
}

// protectedRole returns the AuthRole a field's @protected demands of a
// USER caller. An omitted argument means ADMIN, the schema default.
func protectedRole(field *ast.FieldDefinition) string {
	if field == nil {
		return ""
	}
	directive := field.Directives.ForName(protectedDirective)
	if directive == nil {
		return ""
	}
	argument := directive.Arguments.ForName(roleArgument)
	if argument == nil || argument.Value == nil {
		return roleAdmin
	}
	return argument.Value.Raw
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
