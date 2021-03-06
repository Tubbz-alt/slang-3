package elem

import (
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_MetaStore__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocMetaStore := getBuiltinCfg(metaStoreId)
	a.NotNil(ocMetaStore)
}

func Test_MetaStore__Single(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: metaStoreId,
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "string",
				},
			},
		},
	)
	require.NoError(t, err)
	o.Start()

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1"}, querySrv.Out().Pull())

	o.Main().In().Push("test2")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1", "test2"}, querySrv.Out().Pull())

	o.Main().In().Push("test3")
	o.Main().In().Push("test4")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1", "test2", "test3", "test4"}, querySrv.Out().Pull())
}

func Test_MetaStore__Stream(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: metaStoreId,
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "number",
					},
				},
			},
		},
	)
	require.NoError(t, err)
	o.Start()

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{core.PHMultiple}}, querySrv.Out().Pull())

	o.Main().In().Stream().Push(1.0)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, core.PHMultiple}}, querySrv.Out().Pull())

	o.Main().In().Stream().Push(2.0)
	o.Main().In().PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}}, querySrv.Out().Pull())

	o.Main().In().PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}, []interface{}{core.PHMultiple}}, querySrv.Out().Pull())

	o.Main().In().PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}, []interface{}{}}, querySrv.Out().Pull())
}

func Test_MetaStore__Map(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: metaStoreId,
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "string",
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)
	o.Start()

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Map("a").Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PHSingle,
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").Push(true)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": true,
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("a").Push("test2")
	o.Main().In().Map("b").Push(false)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": true,
		},
		map[string]interface{}{
			"a": "test2",
			"b": false,
		},
	}, querySrv.Out().Pull())
}

func Test_MetaStore__StreamMap(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: metaStoreId,
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"a": {
								Type: "string",
							},
							"b": {
								Type: "boolean",
							},
							"c": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "trigger",
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)
	o.Start()

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("a").Push(o.Main().In().NewBOS())
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{core.PHMultiple}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("a").Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PHSingle,
			"c": core.PHSingle,
		},
		core.PHMultiple,
	}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("b").Push(o.Main().In().NewBOS())
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PHSingle,
			"c": core.PHSingle,
		},
		core.PHMultiple,
	}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("c").Push(o.Main().In().NewBOS())
	o.Main().In().Stream().Map("c").Push(o.Main().In().Stream().Map("c").NewBOS())
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PHSingle,
			"c": []interface{}{core.PHMultiple},
		},
		core.PHMultiple,
	}}, querySrv.Out().Pull())
}

func Test_MetaStore__MapStream(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: metaStoreId,
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "string",
							},
						},
						"b": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "boolean",
							},
						},
						"c": {
							Type: "map",
							Map: map[string]*core.TypeDef{
								"a": {
									Type: "trigger",
								},
								"d": {
									Type: "number",
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)
	o.Start()

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Map("b").PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{core.PHMultiple},
			"c": map[string]interface{}{
				"a": core.PHSingle,
				"d": core.PHSingle,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").Stream().Push(true)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{true, core.PHMultiple},
			"c": map[string]interface{}{
				"a": core.PHSingle,
				"d": core.PHSingle,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("c").Map("a").Push(nil)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{true, core.PHMultiple},
			"c": map[string]interface{}{
				"a": nil,
				"d": core.PHSingle,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("c").Map("a").Push(nil)
	o.Main().In().Map("b").PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{true},
			"c": map[string]interface{}{
				"a": nil,
				"d": core.PHSingle,
			},
		},
		map[string]interface{}{
			"a": core.PHSingle,
			"b": core.PHSingle,
			"c": map[string]interface{}{
				"a": nil,
				"d": core.PHSingle,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{true},
			"c": map[string]interface{}{
				"a": nil,
				"d": core.PHSingle,
			},
		},
		map[string]interface{}{
			"a": core.PHSingle,
			"b": []interface{}{core.PHMultiple},
			"c": map[string]interface{}{
				"a": nil,
				"d": core.PHSingle,
			},
		},
	}, querySrv.Out().Pull())
}
