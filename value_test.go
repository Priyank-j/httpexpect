package httpexpect

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValue_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newValue(chain, nil)
	value.chain.assertFailed(t)

	value.Path("$").chain.assertFailed(t)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(target)

	value.Object().chain.assertFailed(t)
	value.Array().chain.assertFailed(t)
	value.String().chain.assertFailed(t)
	value.Number().chain.assertFailed(t)
	value.Boolean().chain.assertFailed(t)

	value.IsNull()
	value.NotNull()
	value.IsObject()
	value.NotObject()
	value.IsArray()
	value.NotArray()
	value.IsString()
	value.NotString()
	value.IsNumber()
	value.NotNumber()
	value.IsBoolean()
	value.NotBoolean()
	value.IsEqual(nil)
	value.NotEqual(nil)
	value.InList(nil)
	value.NotInList(nil)
}

func TestValue_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewValue(reporter, "Test")
		value.IsEqual("Test")
		value.chain.assertNotFailed(t)
		value.String().chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewValueC(Config{
			Reporter: reporter,
		}, "Test")
		value.IsEqual("Test")
		value.chain.assertNotFailed(t)
		value.String().chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newValue(chain, "Test")
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestValue_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123.0)

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 123.0, target)
	})

	t.Run("target is struct", func(t *testing.T) {
		reporter := newMockReporter(t)

		type S struct {
			Foo int             `json:"foo"`
			Bar []interface{}   `json:"bar"`
			Baz struct{ A int } `json:"baz"`
		}

		m := map[string]interface{}{
			"foo": 123,
			"bar": []interface{}{"123", 456.0},
			"baz": struct{ A int }{123},
		}

		value := NewValue(reporter, m)

		actualStruct := S{
			123,
			[]interface{}{"123", 456.0},
			struct{ A int }{123},
		}

		var target S
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, target, actualStruct)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123)

		value.Decode(nil)

		value.chain.failed()
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewValue(reporter, 123)

		value.Decode(123)

		value.chain.failed()
	})
}

func TestValue_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewValue(reporter, 123)
	assert.Equal(t, []string{"Value()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Value()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Value()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Number()
	assert.Equal(t, []string{"Value()", "Number()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Number()"}, childValue.chain.context.AliasedPath)
}

func TestValue_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("null", func(t *testing.T) {
		var data interface{}

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertFailed(t)
		NewValue(reporter, data).IsNull().chain.assertNotFailed(t)
	})

	t.Run("indirect null", func(t *testing.T) {
		var data []interface{}

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertFailed(t)
		NewValue(reporter, data).IsNull().chain.assertNotFailed(t)
	})

	t.Run("bad", func(t *testing.T) {
		data := func() {}

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})

	t.Run("object", func(t *testing.T) {
		data := map[string]interface{}{}

		NewValue(reporter, data).Object().chain.assertNotFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertNotFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})

	t.Run("array", func(t *testing.T) {
		data := []interface{}{}

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertNotFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertNotFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})

	t.Run("string", func(t *testing.T) {
		data := ""

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertNotFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertNotFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})

	t.Run("number", func(t *testing.T) {
		data := 0.0

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertNotFailed(t)
		NewValue(reporter, data).Boolean().chain.assertFailed(t)
		NewValue(reporter, data).NotNull().chain.assertNotFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})

	t.Run("boolean", func(t *testing.T) {
		data := false

		NewValue(reporter, data).Object().chain.assertFailed(t)
		NewValue(reporter, data).Array().chain.assertFailed(t)
		NewValue(reporter, data).String().chain.assertFailed(t)
		NewValue(reporter, data).Number().chain.assertFailed(t)
		NewValue(reporter, data).Boolean().chain.assertNotFailed(t)
		NewValue(reporter, data).NotNull().chain.assertNotFailed(t)
		NewValue(reporter, data).IsNull().chain.assertFailed(t)
	})
}

func TestValue_GetObject(t *testing.T) {
	type myMap map[string]interface{}

	cases := map[string]struct {
		data           interface{}
		fail           bool
		expectedObject map[string]interface{}
	}{
		"map": {
			data:           map[string]interface{}{"foo": 123.0},
			fail:           false,
			expectedObject: map[string]interface{}{"foo": 123.0},
		},
		"myMap": {
			data:           myMap{"foo": 123.0},
			fail:           false,
			expectedObject: map[string]interface{}(myMap{"foo": 123.0}),
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, instance.data)
			inner := value.Object()

			if instance.fail {
				inner.chain.assertNotFailed(t)
			} else {
				inner.chain.assertNotFailed(t)
				assert.Equal(t, instance.expectedObject, inner.Raw())
			}
		})
	}
}

func TestValue_GetArray(t *testing.T) {
	type myArray []interface{}

	cases := map[string]struct {
		data          interface{}
		fail          bool
		expectedArray []interface{}
	}{
		"array": {
			data:          []interface{}{"foo", 123.0},
			fail:          false,
			expectedArray: []interface{}{"foo", 123.0},
		},
		"myArray": {
			data:          myArray{"foo", 123.0},
			fail:          false,
			expectedArray: []interface{}(myArray{"foo", 123.0}),
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, instance.data)
			inner := value.Array()

			if instance.fail {
				value.chain.assertFailed(t)
				inner.chain.assertFailed(t)
			} else {
				value.chain.assertNotFailed(t)
				inner.chain.assertNotFailed(t)
				assert.Equal(t, instance.expectedArray, inner.Raw())
			}
		})
	}
}

func TestValue_GetString(t *testing.T) {
	type myString string

	cases := map[string]struct {
		data           interface{}
		fail           bool
		expectedString string
	}{
		"string": {
			data:           "foo",
			fail:           false,
			expectedString: "foo",
		},
		"myString": {
			data:           myString("foo"),
			fail:           false,
			expectedString: "foo",
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, instance.data)
			inner := value.String()

			if instance.fail {
				value.chain.assertFailed(t)
				inner.chain.assertFailed(t)
			} else {
				value.chain.assertNotFailed(t)
				inner.chain.assertNotFailed(t)
				assert.Equal(t, instance.expectedString, inner.Raw())
			}
		})
	}
}

func TestValue_GetNumber(t *testing.T) {
	type myInt int

	cases := map[string]struct {
		data        interface{}
		fail        bool
		expectedNum float64
	}{
		"float":   {data: 123.0, fail: false, expectedNum: float64(123.0)},
		"integer": {data: 123, fail: false, expectedNum: float64(123)},
		"myInt":   {data: myInt(123), fail: false, expectedNum: float64(myInt(123))},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, instance.data)
			inner := value.Number()

			if instance.fail {
				value.chain.assertFailed(t)
				inner.chain.assertFailed(t)
			} else {
				value.chain.assertNotFailed(t)
				inner.chain.assertNotFailed(t)
				assert.Equal(t, instance.expectedNum, inner.Raw())
			}
		})
	}
}

func TestValue_GetBoolean(t *testing.T) {
	type myBool bool

	cases := map[string]struct {
		data         interface{}
		fail         bool
		expectedBool bool
	}{
		"false":  {data: false, fail: false, expectedBool: false},
		"true":   {data: true, fail: false, expectedBool: true},
		"myTrue": {data: myBool(true), fail: false, expectedBool: true},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			value := NewValue(reporter, instance.data)
			inner := value.Boolean()

			if instance.fail {
				value.chain.assertFailed(t)
				inner.chain.assertFailed(t)
			} else {
				value.chain.assertNotFailed(t)
				inner.chain.assertNotFailed(t)
				assert.Equal(t, instance.expectedBool, inner.Raw())
			}
		})
	}
}

func TestValue_IsObject(t *testing.T) {
	cases := map[string]struct {
		data       interface{}
		wantObject bool
	}{
		"object": {data: map[string]interface{}{"foo": 123.0}, wantObject: true},
		"string": {data: "foo", wantObject: false},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.wantObject {
				NewValue(reporter, instance.data).IsObject().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).NotObject().
					chain.assertFailed(t)
			} else {
				NewValue(reporter, instance.data).NotObject().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).IsObject().
					chain.assertFailed(t)
			}
		})
	}
}

func TestValue_IsArray(t *testing.T) {
	cases := map[string]struct {
		data      interface{}
		wantArray bool
	}{
		"array":  {data: []interface{}{"foo", "123"}, wantArray: true},
		"string": {data: "foo", wantArray: false},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.wantArray {
				NewValue(reporter, instance.data).IsArray().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).NotArray().
					chain.assertFailed(t)
			} else {
				NewValue(reporter, instance.data).NotArray().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).IsArray().
					chain.assertFailed(t)
			}
		})
	}
}

func TestValue_IsString(t *testing.T) {
	cases := map[string]struct {
		data       interface{}
		wantString bool
	}{
		"string":  {data: "foo", wantString: true},
		"integer": {data: 123, wantString: false},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.wantString {
				NewValue(reporter, instance.data).IsString().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).NotString().
					chain.assertFailed(t)
			} else {
				NewValue(reporter, instance.data).NotString().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).IsString().
					chain.assertFailed(t)
			}
		})
	}
}

func TestValue_IsNumber(t *testing.T) {
	cases := map[string]struct {
		data       interface{}
		wantNumber bool
	}{
		"integer": {data: 123, wantNumber: true},
		"string":  {data: "foo", wantNumber: false},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.wantNumber {
				NewValue(reporter, instance.data).IsNumber().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).NotNumber().
					chain.assertFailed(t)
			} else {
				NewValue(reporter, instance.data).NotNumber().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).IsNumber().
					chain.assertFailed(t)
			}
		})
	}
}

func TestValue_IsBoolean(t *testing.T) {
	cases := map[string]struct {
		data     interface{}
		wantBool bool
	}{
		"bool":   {data: true, wantBool: true},
		"string": {data: "foo", wantBool: false},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.wantBool {
				NewValue(reporter, instance.data).IsBoolean().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).NotBoolean().
					chain.assertFailed(t)
			} else {
				NewValue(reporter, instance.data).NotBoolean().
					chain.assertNotFailed(t)

				NewValue(reporter, instance.data).IsBoolean().
					chain.assertFailed(t)
			}
		})
	}
}

func TestValue_IsEqual(t *testing.T) {
	reporter := newMockReporter(t)

	data1 := map[string]interface{}{"foo": "bar"}
	data2 := "baz"

	NewValue(reporter, data1).IsEqual(data1).chain.assertNotFailed(t)
	NewValue(reporter, data2).IsEqual(data2).chain.assertNotFailed(t)

	NewValue(reporter, data1).NotEqual(data1).chain.assertFailed(t)
	NewValue(reporter, data2).NotEqual(data2).chain.assertFailed(t)

	NewValue(reporter, data1).IsEqual(data2).chain.assertFailed(t)
	NewValue(reporter, data2).IsEqual(data1).chain.assertFailed(t)

	NewValue(reporter, data1).NotEqual(data2).chain.assertNotFailed(t)
	NewValue(reporter, data2).NotEqual(data1).chain.assertNotFailed(t)

	NewValue(reporter, nil).IsEqual(nil).chain.assertNotFailed(t)

	NewValue(reporter, nil).IsEqual(map[string]interface{}(nil)).chain.assertNotFailed(t)
	NewValue(reporter, nil).IsEqual(map[string]interface{}{}).chain.assertFailed(t)

	NewValue(reporter, data1).IsEqual(func() {}).chain.assertFailed(t)
	NewValue(reporter, data1).NotEqual(func() {}).chain.assertFailed(t)
}

func TestValue_InList(t *testing.T) {
	reporter := newMockReporter(t)

	data1 := map[string]interface{}{"foo": "bar"}
	data2 := "baz"
	data3 := struct {
		Data []int `json:"data"`
	}{
		Data: []int{1, 2, 3, 4},
	}

	NewValue(reporter, data1).InList().chain.assertFailed(t)
	NewValue(reporter, data2).NotInList().chain.assertFailed(t)

	NewValue(reporter, data1).InList(data1, data3).chain.assertNotFailed(t)
	NewValue(reporter, data2).NotInList(data1, data3).chain.assertNotFailed(t)

	NewValue(reporter, data1).InList(data2, data3).chain.assertFailed(t)
	NewValue(reporter, data2).NotInList(data2, data3).chain.assertFailed(t)

	NewValue(reporter, data1).InList(data2).chain.assertFailed(t)
	NewValue(reporter, data2).NotInList(data2).chain.assertFailed(t)

	NewValue(reporter, data1).InList(data1).chain.assertNotFailed(t)
	NewValue(reporter, data2).NotInList(data1).chain.assertNotFailed(t)

	NewValue(reporter, nil).InList(map[string]interface{}(nil)).chain.assertNotFailed(t)
	NewValue(reporter, nil).NotInList(map[string]interface{}{}).chain.assertNotFailed(t)

	NewValue(reporter, data1).InList(func() {}).chain.assertFailed(t)
	NewValue(reporter, data1).NotInList(func() {}).chain.assertFailed(t)

	NewValue(reporter, data1).InList(data1, func() {}).chain.assertFailed(t)
	NewValue(reporter, data1).NotInList(data1, func() {}).chain.assertFailed(t)

	NewValue(reporter, data1).InList(data2, func() {}).chain.assertFailed(t)
	NewValue(reporter, data1).NotInList(data2, func() {}).chain.assertFailed(t)
}

func TestValue_PathTypes(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("object", func(t *testing.T) {
		user0 := map[string]interface{}{"name": "john"}
		user1 := map[string]interface{}{"name": "bob"}

		data := map[string]interface{}{
			"users": []interface{}{
				user0,
				user1,
			},
		}

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		assert.Equal(t, data["users"], value.Path("$.users").Raw())
		assert.Equal(t, user0, value.Path("$.users[0]").Raw())
		assert.Equal(t, "john", value.Path("$.users[0].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$.users[*].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$..name").Raw())
		value.chain.assertNotFailed(t)

		names := value.Path("$..name").Array().Iter()
		names[0].String().IsEqual("john").chain.assertNotFailed(t)
		names[1].String().IsEqual("bob").chain.assertNotFailed(t)
		value.chain.assertNotFailed(t)

		for _, key := range []string{"$.bad", "!"} {
			bad := value.Path(key)
			assert.True(t, bad != nil)
			assert.True(t, bad.Raw() == nil)
			value.chain.assertFailed(t)
			value.chain.clearFailed()
		}
	})

	t.Run("array", func(t *testing.T) {
		user0 := map[string]interface{}{"name": "john"}
		user1 := map[string]interface{}{"name": "bob"}

		data := []interface{}{
			user0,
			user1,
		}

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		assert.Equal(t, user0, value.Path("$[0]").Raw())
		assert.Equal(t, "john", value.Path("$[0].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$[*].name").Raw())
		assert.Equal(t, []interface{}{"john", "bob"}, value.Path("$..name").Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("string", func(t *testing.T) {
		data := "foo"

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("number", func(t *testing.T) {
		data := 123

		value := NewValue(reporter, data)

		assert.Equal(t, float64(data), value.Path("$").Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("boolean", func(t *testing.T) {
		data := true

		value := NewValue(reporter, data)

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("null", func(t *testing.T) {
		value := NewValue(reporter, nil)

		assert.Equal(t, nil, value.Path("$").Raw())
		value.chain.assertNotFailed(t)
	})

	t.Run("error", func(t *testing.T) {
		data := "foo"

		value := NewValue(reporter, data)

		for _, key := range []string{"$.bad", "!"} {
			bad := value.Path(key)
			assert.True(t, bad != nil)
			assert.True(t, bad.Raw() == nil)
			value.chain.assertFailed(t)
		}
	})

	t.Run("int float", func(t *testing.T) {
		data := map[string]interface{}{
			"A": 123,
			"B": 123.0,
		}

		value := NewValue(reporter, data)
		value.chain.assertNotFailed(t)

		a := value.Path(`$["A"]`)
		a.chain.assertNotFailed(t)
		assert.Equal(t, 123.0, a.Raw())

		b := value.Path(`$["B"]`)
		b.chain.assertNotFailed(t)
		assert.Equal(t, 123.0, b.Raw())
	})
}

// based on github.com/yalp/jsonpath
func TestValue_PathExpressions(t *testing.T) {
	data := map[string]interface{}{
		"A": []interface{}{
			"string",
			23.3,
			3.0,
			true,
			false,
			nil,
		},
		"B": "value",
		"C": 3.14,
		"D": map[string]interface{}{
			"C": 3.1415,
			"V": []interface{}{
				"string2a",
				"string2b",
				map[string]interface{}{
					"C": 3.141592,
				},
			},
		},
		"E": map[string]interface{}{
			"A": []interface{}{"string3"},
			"D": map[string]interface{}{
				"V": map[string]interface{}{
					"C": 3.14159265,
				},
			},
		},
		"F": map[string]interface{}{
			"V": []interface{}{
				"string4a",
				"string4b",
				map[string]interface{}{
					"CC": 3.1415926535,
				},
				map[string]interface{}{
					"CC": "hello",
				},
				[]interface{}{
					"string5a",
					"string5b",
				},
				[]interface{}{
					"string6a",
					"string6b",
				},
			},
		},
	}

	reporter := newMockReporter(t)

	runTests := func(tests map[string]interface{}) {
		value := NewValue(reporter, data)
		value.chain.assertNotFailed(t)

		for path, expected := range tests {
			actual := value.Path(path)
			actual.chain.assertNotFailed(t)

			assert.Equal(t, expected, actual.Raw())
		}
	}

	t.Run("pick", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$":         data,
			"$.A[0]":    "string",
			`$["A"][0]`: "string",
			"$.A":       []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A[*]":    []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*":     []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*.a":   []interface{}{},
		})
	})

	t.Run("slice", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$.A[1,4,2]":      []interface{}{23.3, false, 3.0},
			`$["B","C"]`:      []interface{}{"value", 3.14},
			`$["C","B"]`:      []interface{}{3.14, "value"},
			"$.A[1:4]":        []interface{}{23.3, 3.0, true},
			"$.A[::2]":        []interface{}{"string", 3.0, false},
			"$.A[-2:]":        []interface{}{false, nil},
			"$.A[:-1]":        []interface{}{"string", 23.3, 3.0, true, false},
			"$.A[::-1]":       []interface{}{nil, false, true, 3.0, 23.3, "string"},
			"$.F.V[4:5][0,1]": []interface{}{"string5a", "string5b"},
			"$.F.V[4:6][1]":   []interface{}{"string5b", "string6b"},
			"$.F.V[4:6][0,1]": []interface{}{"string5a", "string5b", "string6a", "string6b"},
			"$.F.V[4,5][0:2]": []interface{}{"string5a", "string5b", "string6a", "string6b"},
			"$.F.V[4:6]": []interface{}{
				[]interface{}{
					"string5a",
					"string5b",
				},
				[]interface{}{
					"string6a",
					"string6b",
				},
			},
		})
	})

	t.Run("quote", func(t *testing.T) {
		runTests(map[string]interface{}{
			`$[A][0]`:    "string",
			`$["A"][0]`:  "string",
			`$[B,C]`:     []interface{}{"value", 3.14},
			`$["B","C"]`: []interface{}{"value", 3.14},
		})
	})

	t.Run("search", func(t *testing.T) {
		runTests(map[string]interface{}{
			"$..C":       []interface{}{3.14, 3.1415, 3.141592, 3.14159265},
			`$..["C"]`:   []interface{}{3.14, 3.1415, 3.141592, 3.14159265},
			"$.D.V..C":   []interface{}{3.141592},
			"$.D.V.*.C":  []interface{}{3.141592},
			"$.D.V..*.C": []interface{}{3.141592},
			"$.D.*..C":   []interface{}{3.141592},
			"$.*.V..C":   []interface{}{3.141592},
			"$.*.D.V.C":  []interface{}{3.14159265},
			"$.*.D..C":   []interface{}{3.14159265},
			"$.*.D.V..*": []interface{}{3.14159265},
			"$..D..V..C": []interface{}{3.141592, 3.14159265},
			"$.*.*.*.C":  []interface{}{3.141592, 3.14159265},
			"$..V..C":    []interface{}{3.141592, 3.14159265},
			"$.D.V..*": []interface{}{
				"string2a",
				"string2b",
				map[string]interface{}{
					"C": 3.141592,
				},
				3.141592,
			},
			"$..A": []interface{}{
				[]interface{}{"string", 23.3, 3.0, true, false, nil},
				[]interface{}{"string3"},
			},
			"$..A..*":      []interface{}{"string", 23.3, 3.0, true, false, nil, "string3"},
			"$.A..*":       []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$.A.*":        []interface{}{"string", 23.3, 3.0, true, false, nil},
			"$..A[0,1]":    []interface{}{"string", 23.3},
			"$..A[0]":      []interface{}{"string", "string3"},
			"$.*.V[0]":     []interface{}{"string2a", "string4a"},
			"$.*.V[1]":     []interface{}{"string2b", "string4b"},
			"$.*.V[0,1]":   []interface{}{"string2a", "string2b", "string4a", "string4b"},
			"$.*.V[0:2]":   []interface{}{"string2a", "string2b", "string4a", "string4b"},
			"$.*.V[2].C":   []interface{}{3.141592},
			"$..V[2].C":    []interface{}{3.141592},
			"$..V[*].C":    []interface{}{3.141592},
			"$.*.V[2].*":   []interface{}{3.141592, 3.1415926535},
			"$.*.V[2:3].*": []interface{}{3.141592, 3.1415926535},
			"$.*.V[2:4].*": []interface{}{3.141592, 3.1415926535, "hello"},
			"$..V[2,3].CC": []interface{}{3.1415926535, "hello"},
			"$..V[2:4].CC": []interface{}{3.1415926535, "hello"},
			"$..V[*].*": []interface{}{
				3.141592,
				3.1415926535,
				"hello",
				"string5a",
				"string5b",
				"string6a",
				"string6b",
			},
			"$..[0]": []interface{}{
				"string",
				"string2a",
				"string3",
				"string4a",
				"string5a",
				"string6a",
			},
			"$..ZZ": []interface{}{},
		})
	})
}

func TestValue_Schema(t *testing.T) {
	reporter := newMockReporter(t)

	schema := `{
		"type": "object",
		"properties": {
			"foo": {
				"type": "string"
			},
			"bar": {
				"type": "integer"
			}
		},
		"require": ["foo", "bar"]
	}`

	data1 := map[string]interface{}{
		"foo": "a",
		"bar": 1,
	}

	data2 := map[string]interface{}{
		"foo": "a",
		"bar": "b",
	}

	NewValue(reporter, data1).Schema(schema).chain.assertNotFailed(t)
	NewValue(reporter, data2).Schema(schema).chain.assertFailed(t)

	NewValue(reporter, data1).Schema([]byte(schema)).chain.assertNotFailed(t)
	NewValue(reporter, data2).Schema([]byte(schema)).chain.assertFailed(t)

	var b interface{}
	err := json.Unmarshal([]byte(schema), &b)
	require.Nil(t, err)

	NewValue(reporter, data1).Schema(b).chain.assertNotFailed(t)
	NewValue(reporter, data2).Schema(b).chain.assertFailed(t)

	tmp, _ := ioutil.TempFile("", "httpexpect")
	defer os.Remove(tmp.Name())

	_, err = tmp.Write([]byte(schema))
	require.Nil(t, err)

	err = tmp.Close()
	require.Nil(t, err)

	url := "file://" + tmp.Name()

	NewValue(reporter, data1).Schema(url).chain.assertNotFailed(t)
	NewValue(reporter, data2).Schema(url).chain.assertFailed(t)

	NewValue(reporter, data1).Schema("file:///bad/path").chain.assertFailed(t)
	NewValue(reporter, data1).Schema("{ bad json").chain.assertFailed(t)
}
