package graph

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
)

// TestRootFieldsAreClassified is the test the operation gate's design calls
// for: a root field carrying neither directive fails the build, rather than
// silently becoming reachable or silently becoming protected.
//
// The gate itself only looks for @public, so @protected buys nothing at
// runtime. It exists exactly so that this test can tell "decided to protect"
// apart from "never thought about it".
func TestRootFieldsAreClassified(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	check := func(defName string, def *ast.Definition) {
		t.Helper()
		if def == nil {
			return
		}
		for _, field := range def.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			isPublic := field.Directives.ForName(publicDirective) != nil
			isProtected := field.Directives.ForName(protectedDirective) != nil
			if isPublic && isProtected {
				t.Errorf("%s.%s is marked both @public and @protected", defName, field.Name)
			}
			if !isPublic && !isProtected {
				t.Errorf("%s.%s is marked neither @public nor @protected", defName, field.Name)
			}

		}
	}

	check("Query", schema.Query)
	check("Mutation", schema.Mutation)
}

// TestEveryAuthKindIsEnforced catches an AuthKind added to the schema that
// the gate has no case for. protectedKinds would return it, nothing would
// match it, and a field naming it would become unreachable by anyone.
func TestEveryAuthKindIsEnforced(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	kind := schema.Types["AuthKind"]
	if kind == nil {
		t.Fatal("the AuthKind enum is gone: the @protected kinds argument has no vocabulary")
	}
	enforced := map[string]bool{kindUser: true, kindDisplay: true}
	for _, value := range kind.EnumValues {
		if !enforced[value.Name] {
			t.Errorf("AuthKind.%s has no case in the gate", value.Name)
		}
	}
	// The gate matches on strings pulled from the AST, so it is worth
	// confirming the generated constants it compares against are really
	// values of the schema enum and not stale names that match nothing.
	for name := range enforced {
		if kind.EnumValues.ForName(name) == nil {
			t.Errorf("the gate enforces %q, which AuthKind does not define", name)
		}
	}
}

// TestEveryAuthRoleIsEnforced is TestEveryAuthKindIsEnforced's counterpart
// for the role argument: a role added to the schema with no case here would
// silently never match a USER caller, locking every field that named it.
func TestEveryAuthRoleIsEnforced(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	role := schema.Types["AuthRole"]
	if role == nil {
		t.Fatal("the AuthRole enum is gone: the @protected role argument has no vocabulary")
	}
	enforced := map[string]bool{roleAdmin: true}
	for _, value := range role.EnumValues {
		if !enforced[value.Name] {
			t.Errorf("AuthRole.%s has no case in the gate", value.Name)
		}
	}
	for name := range enforced {
		if role.EnumValues.ForName(name) == nil {
			t.Errorf("the gate enforces %q, which AuthRole does not define", name)
		}
	}
}

// TestDisplayOnlyFieldsDoNotSetRole catches a field combining kinds that
// exclude USER with a role argument - a combination that can never be
// satisfied, since role is only ever checked for a USER caller.
func TestDisplayOnlyFieldsDoNotSetRole(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	check := func(defName string, def *ast.Definition) {
		t.Helper()
		if def == nil {
			return
		}
		for _, field := range def.Fields {
			kinds := protectedKinds(field)
			if len(kinds) == 0 || containsString(kinds, kindUser) {
				continue
			}
			directive := field.Directives.ForName(protectedDirective)
			if directive.Arguments.ForName(roleArgument) != nil {
				t.Errorf("%s.%s sets role but its kinds exclude USER: role is never checked for it", defName, field.Name)
			}
		}
	}

	check("Query", schema.Query)
	check("Mutation", schema.Mutation)
}

// TestOnlyIntendedRootFieldsArePublic pins the public surface by name. The
// directive test above only checks that a decision was made; this one fails
// if that decision changes, so widening the unauthenticated surface has to
// be deliberate rather than a one-character schema edit.
func TestOnlyIntendedRootFieldsArePublic(t *testing.T) {
	schema := NewExecutableSchema(Config{Resolvers: &Resolver{}}).Schema()

	want := map[string]bool{
		"Query.health":   true,
		"Query.session":  true,
		"Mutation.login": true,
		"Mutation.setup": true,
	}

	got := map[string]bool{}
	for defName, def := range map[string]*ast.Definition{"Query": schema.Query, "Mutation": schema.Mutation} {
		if def == nil {
			continue
		}
		for _, field := range def.Fields {
			if strings.HasPrefix(field.Name, "__") {
				continue
			}
			if field.Directives.ForName(publicDirective) != nil {
				got[defName+"."+field.Name] = true
			}
		}
	}

	for name := range got {
		if !want[name] {
			t.Errorf("%s is newly @public: widening the unauthenticated surface needs a deliberate change here", name)
		}
	}
	for name := range want {
		if !got[name] {
			t.Errorf("%s is no longer @public", name)
		}
	}
}
