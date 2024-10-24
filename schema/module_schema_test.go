package schema

import (
	"reflect"
	"strings"
	"testing"
)

func TestModuleSchema_Validate(t *testing.T) {
	tests := []struct {
		name        string
		objectTypes []ObjectType
		errContains string
	}{
		{
			name: "valid module schema",
			objectTypes: []ObjectType{
				{
					Name: "object1",
					KeyFields: []Field{
						{
							Name: "field1",
							Kind: StringKind,
						},
					},
				},
			},
			errContains: "",
		},
		{
			name: "invalid object type",
			objectTypes: []ObjectType{
				{
					Name: "",
					KeyFields: []Field{
						{
							Name: "field1",
							Kind: StringKind,
						},
					},
				},
			},
			errContains: "invalid object type name",
		},
		{
			name: "same enum with missing values",
			objectTypes: []ObjectType{
				{
					Name: "object1",
					KeyFields: []Field{
						{
							Name: "k",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "b"},
							},
						},
					},
					ValueFields: []Field{
						{
							Name: "v",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "b", "c"},
							},
						},
					},
				},
			},
			errContains: "different number of values",
		},
		{
			name: "same enum with different values",
			objectTypes: []ObjectType{
				{
					Name: "object1",
					KeyFields: []Field{
						{
							Name: "k",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "b"},
							},
						},
					},
				},
				{
					Name: "object2",
					KeyFields: []Field{
						{
							Name: "k",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "c"},
							},
						},
					},
				},
			},
			errContains: "different values",
		},
		{
			name: "same enum",
			objectTypes: []ObjectType{{
				Name: "object1",
				KeyFields: []Field{
					{
						Name: "k",
						Kind: EnumKind,
						EnumType: EnumType{
							Name:   "enum1",
							Values: []string{"a", "b"},
						},
					},
				},
			},
				{
					Name: "object2",
					KeyFields: []Field{
						{
							Name: "k",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "b"},
							},
						},
					},
				},
			},
		},
		{
			objectTypes: []ObjectType{
				{
					Name: "type1",
					ValueFields: []Field{
						{
							Name: "field1",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "type1",
								Values: []string{"a", "b"},
							},
						},
					},
				},
			},
			errContains: "enum \"type1\" already exists as a different non-enum type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// because validate is called when calling NewModuleSchema, we just call NewModuleSchema
			_, err := NewModuleSchema(tt.objectTypes)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestModuleSchema_ValidateObjectUpdate(t *testing.T) {
	tests := []struct {
		name         string
		moduleSchema ModuleSchema
		objectUpdate ObjectUpdate
		errContains  string
	}{
		{
			name: "valid object update",
			moduleSchema: requireModuleSchema(t, []ObjectType{
				{
					Name: "object1",
					KeyFields: []Field{
						{
							Name: "field1",
							Kind: StringKind,
						},
					},
				},
			},
			),
			objectUpdate: ObjectUpdate{
				TypeName: "object1",
				Key:      "abc",
			},
			errContains: "",
		},
		{
			name: "object type not found",
			moduleSchema: requireModuleSchema(t, []ObjectType{
				{
					Name: "object1",
					KeyFields: []Field{
						{
							Name: "field1",
							Kind: StringKind,
						},
					},
				},
			},
			),
			objectUpdate: ObjectUpdate{
				TypeName: "object2",
				Key:      "abc",
			},
			errContains: "object type \"object2\" not found in module schema",
		},
		{
			name: "type name refers to an enum",
			moduleSchema: requireModuleSchema(t, []ObjectType{
				{
					Name: "obj1",
					KeyFields: []Field{
						{
							Name: "field1",
							Kind: EnumKind,
							EnumType: EnumType{
								Name:   "enum1",
								Values: []string{"a", "b"},
							},
						},
					},
				},
			}),
			objectUpdate: ObjectUpdate{
				TypeName: "enum1",
				Key:      "a",
			},
			errContains: "type \"enum1\" is not an object type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.moduleSchema.ValidateObjectUpdate(tt.objectUpdate)
			if tt.errContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error to contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func requireModuleSchema(t *testing.T, objectTypes []ObjectType) ModuleSchema {
	t.Helper()
	moduleSchema, err := NewModuleSchema(objectTypes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return moduleSchema
}

func TestModuleSchema_LookupType(t *testing.T) {
	moduleSchema := requireModuleSchema(t, []ObjectType{
		{
			Name: "object1",
			KeyFields: []Field{
				{
					Name: "field1",
					Kind: StringKind,
				},
			},
		},
	})

	typ, ok := moduleSchema.LookupType("object1")
	if !ok {
		t.Fatalf("expected to find object type \"object1\"")
	}

	objectType, ok := typ.(ObjectType)
	if !ok {
		t.Fatalf("expected object type, got %T", typ)
	}

	if objectType.Name != "object1" {
		t.Fatalf("expected object type name \"object1\", got %q", objectType.Name)
	}
}

func exampleSchema(t *testing.T) ModuleSchema {
	return requireModuleSchema(t, []ObjectType{
		{
			Name: "object1",
			KeyFields: []Field{
				{
					Name: "field1",
					Kind: EnumKind,
					EnumType: EnumType{
						Name:   "enum2",
						Values: []string{"d", "e", "f"},
					},
				},
			},
		},
		{
			Name: "object2",
			KeyFields: []Field{
				{
					Name: "field1",
					Kind: EnumKind,
					EnumType: EnumType{
						Name:   "enum1",
						Values: []string{"a", "b", "c"},
					},
				},
			},
		},
	})
}

func TestModuleSchema_Types(t *testing.T) {
	moduleSchema := exampleSchema(t)

	var typeNames []string
	moduleSchema.Types(func(typ Type) bool {
		typeNames = append(typeNames, typ.TypeName())
		return true
	})

	expected := []string{"enum1", "enum2", "object1", "object2"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}

	typeNames = nil
	// scan just the first type and return false
	moduleSchema.Types(func(typ Type) bool {
		typeNames = append(typeNames, typ.TypeName())
		return false
	})

	expected = []string{"enum1"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}
}

func TestModuleSchema_ObjectTypes(t *testing.T) {
	moduleSchema := exampleSchema(t)

	var typeNames []string
	moduleSchema.ObjectTypes(func(typ ObjectType) bool {
		typeNames = append(typeNames, typ.Name)
		return true
	})

	expected := []string{"object1", "object2"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}

	typeNames = nil
	// scan just the first type and return false
	moduleSchema.ObjectTypes(func(typ ObjectType) bool {
		typeNames = append(typeNames, typ.Name)
		return false
	})

	expected = []string{"object1"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}
}

func TestModuleSchema_EnumTypes(t *testing.T) {
	moduleSchema := exampleSchema(t)

	var typeNames []string
	moduleSchema.EnumTypes(func(typ EnumType) bool {
		typeNames = append(typeNames, typ.Name)
		return true
	})

	expected := []string{"enum1", "enum2"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}

	typeNames = nil
	// scan just the first type and return false
	moduleSchema.EnumTypes(func(typ EnumType) bool {
		typeNames = append(typeNames, typ.Name)
		return false
	})

	expected = []string{"enum1"}
	if !reflect.DeepEqual(typeNames, expected) {
		t.Fatalf("expected %v, got %v", expected, typeNames)
	}
}
