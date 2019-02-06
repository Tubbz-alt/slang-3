package tests

import (
	"github.com/Bitspark/slang/tests/assertions"
	"io/ioutil"
	"testing"
)

func TestTestOperator__TrivialTests(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/voidOp.json", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(1, succs)
	a.Equal(0, fails)
}

func TestTestOperator__SimpleFail(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/voidOp_corruptTest.json", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(0, succs)
	a.Equal(1, fails)
}

func TestTestOperator__ComplexTest(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/nested_op/usingSubCustomOpDouble.json", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(2, succs)
	a.Equal(0, fails)
}

func TestTestOperator__SuiteTests(t *testing.T) {
	a := assertions.New(t)

	succs, fails, err := Test.RunTestBench("test_data/suite/polynomial.yaml", ioutil.Discard, false)
	a.NoError(err)
	a.Equal(1, succs)
	a.Equal(0, fails)

	succs, fails, err = Test.RunTestBench("test_data/suite/main.yaml", ioutil.Discard, false)
	a.NoError(err)
	a.Equal(2, succs)
	a.Equal(0, fails)
}

func TestOperator_Pack(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/slib/pack.yaml", ioutil.Discard, false)
	a.NoError(err)
	a.Equal(1, succs)
	a.Equal(0, fails)
}

func TestTestOperator__SumReduce(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/sum/reduce.yaml", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(4, succs)
	a.Equal(0, fails)
}

func TestTestOperator__MergeSort(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/slib/merge_sort.yaml", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(5, succs)
	a.Equal(0, fails)
}

func TestTestOperator_Properties(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/properties/prop_op.yaml", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(3, succs)
	a.Equal(0, fails)
}

func TestTestOperator_NestedProperties(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/properties/prop2_op.yaml", ioutil.Discard, false)
	a.NoError(err)
	a.Equal(3, succs)
	a.Equal(0, fails)
}

func TestTestOperator_NestedDelegates(t *testing.T) {
	a := assertions.New(t)
	succs, fails, err := Test.RunTestBench("test_data/delegates/wrapper_op.yaml", ioutil.Discard, true)
	a.NoError(err)
	a.Equal(3, succs)
	a.Equal(0, fails)
}
