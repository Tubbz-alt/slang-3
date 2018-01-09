package tests

import (
	"testing"
	"slang"
	"slang/tests/assertions"
	"slang/core"
	"github.com/stretchr/testify/require"
)

// OPERATOR DEFINITION

func TestOperatorDef_SpecifyGenericPorts__NilGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "number"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(nil))
}

func TestOperatorDef_SpecifyGenericPorts__InPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "generic", "generic": "g1"}, "out": {"type": "number"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.In.Type)
}

func TestOperatorDef_SpecifyGenericPorts__OutPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "generic", "generic": "g1"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.Out.Type)
}

func TestOperatorDef_SpecifyGenericPorts__GenericPortsGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(
		`{"in": {"type": "number"}, "out": {"type": "number"}, "operators": [{"name": "test", "operator": "fork", "generics": {"itemType": {"type": "generic", "generic": "g1"}}}]}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g1": {
			Type: "boolean",
		},
	}))
	a.Equal("boolean", op.Operators[0].Generics["itemType"].Type)
}

func TestOperatorDef_SpecifyGenericPorts__DifferentIdentifier(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "generic", "generic": "g1"}, "out": {"type": "number"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.SpecifyGenericPorts(map[string]*core.PortDef{
		"g2": {
			Type: "boolean",
		},
	}))
	a.Equal("generic", op.In.Type)
}

func TestOperatorDef_FreeOfGenerics__InPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "generic", "generic": "t1"}, "out": {"type": "number"}}`)
	require.NoError(t, op.Validate())
	a.Error(op.FreeOfGenerics())
}

func TestOperatorDef_FreeOfGenerics__InPortNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "string"}, "out": {"type": "number"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.FreeOfGenerics())
}

func TestOperatorDef_FreeOfGenerics__OutPortGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "generic", "generic": "t1"}}`)
	require.NoError(t, op.Validate())
	a.Error(op.FreeOfGenerics())
}

func TestOperatorDef_FreeOfGenerics__OutPortNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(`{"in": {"type": "number"}, "out": {"type": "string"}}`)
	require.NoError(t, op.Validate())
	a.NoError(op.FreeOfGenerics())
}

func TestOperatorDef_FreeOfGenerics__GenericPortsGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(
		`{"in": {"type": "number"}, "out": {"type": "number"}, "operators": [{"name": "test", "operator": "fork", "generics": {"itemType": {"type": "generic", "generic": "g1"}}}]}`)
	require.NoError(t, op.Validate())
	a.Error(op.FreeOfGenerics())
}

func TestOperatorDef_FreeOfGenerics__GenericPortsNoGenerics(t *testing.T) {
	a := assertions.New(t)
	op := slang.ParseOperatorDef(
		`{"in": {"type": "number"}, "out": {"type": "number"}, "operators": [{"name": "test", "operator": "fork", "generics": {"itemType": {"type": "number"}}}]}`)
	require.NoError(t, op.Validate())
	a.NoError(op.FreeOfGenerics())
}

// PORT DEFINITION

func TestPortDef_Copy__Simple(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "number"}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("number", pdCpy.Type)
}

func TestPortDef_Copy__Stream(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "string"}}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("string", pdCpy.Stream.Type)
	a.False(pd.Stream == pdCpy.Stream)
}

func TestPortDef_Copy__Map(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "string"}}}
	require.NoError(t, pd.Validate())
	pdCpy := pd.Copy()
	require.NoError(t, pdCpy.Validate())
	a.Equal("string", pdCpy.Map["a"].Type)
	a.False(pd.Map["a"] == pdCpy.Map["a"])
}

func TestPortDef_SpecifyGenericPorts__Simple(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenericPorts(map[string]*core.PortDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Type)
}

func TestPortDef_SpecifyGenericPorts__DifferentIdentifier(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenericPorts(map[string]*core.PortDef{
		"t2": {Type: "number"},
	}))
	a.Equal("generic", pd.Type)
}

func TestPortDef_SpecifyGenericPorts__MultipleIdentifiers(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{
		Type: "map",
		Map:  map[string]*core.PortDef{"a": {Type: "generic", Generic: "t1"}, "b": {Type: "generic", Generic: "t2"}},
	}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenericPorts(map[string]*core.PortDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Map["a"].Type)
	a.Equal("generic", pd.Map["b"].Type)
}

func TestPortDef_SpecifyGenericPorts__Stream(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "generic", Generic: "t1"}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenericPorts(map[string]*core.PortDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Stream.Type)
}

func TestPortDef_SpecifyGenericPorts__Map(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "generic", Generic: "t1"}}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.SpecifyGenericPorts(map[string]*core.PortDef{
		"t1": {Type: "number"},
	}))
	a.Equal("number", pd.Map["a"].Type)
}

func TestPortDef_FreeOfGenerics__SimpleGeneric(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "generic", Generic: "t1"}
	require.NoError(t, pd.Validate())
	a.Error(pd.FreeOfGenerics())
}

func TestPortDef_FreeOfGenerics__SimpleNoGeneric(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "number"}
	require.NoError(t, pd.Validate())
	a.NoError(pd.FreeOfGenerics())
}

func TestPortDef_FreeOfGenerics__StreamGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "generic", Generic: "t1"}}
	require.NoError(t, pd.Validate())
	a.Error(pd.FreeOfGenerics())
}

func TestPortDef_FreeOfGenerics__StreamNoGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "number"}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.FreeOfGenerics())
}

func TestPortDef_FreeOfGenerics__MapGenerics(t *testing.T) {
	a := assertions.New(t)
	pd := core.PortDef{Type: "map", Map: map[string]*core.PortDef{"a": {Type: "number"}}}
	require.NoError(t, pd.Validate())
	a.NoError(pd.FreeOfGenerics())
}
