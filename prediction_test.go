package main

import (
	"testing"
	"fmt"
	"os"
)

var dataSet [][]float64

var MIN_DATA_POINTS = 12

// Unit Test
func TestLinearRegression(t *testing.T) {
	if (len(dataSet) < MIN_DATA_POINTS) {
		dataSet = loadDataSet("USD", "TRY")
	}
	if (len(dataSet) < MIN_DATA_POINTS) {
		t.Error("expected", MIN_DATA_POINTS, "got", len(dataSet))
	}
	rmse := evaluateAlgorithm(dataSet, simpleLinearRegression)
	fmt.Println(fmt.Sprintf("RMSE: %.3f", rmse))
}

// Integration Test
func TestMain(m *testing.M)  {
	dataSet = loadDataSet("USD", "TRY")
	os.Exit(m.Run())
}
