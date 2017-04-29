package main

import (
	"fmt"
	"math"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"os"
	"errors"
	"log"
)

var (
	errNotEnoughArgs =  errors.New("Not enough argument.")
	errNotEnoughData = errors.New("Not enough data points")
)

type Fixer struct {
	Date string
	Rates interface{}
}

// Load dataset from fixer
func loadDataSet(from, to string) [][]float64 {
	var results [][]float64
	for i := 1; i <= 12; i++ {
		m := ""
		if (i < 10) {
			m = "0"
		}
		mm := fmt.Sprintf("%s%d", m, i)
		resp, err := http.Get(fmt.Sprintf("http://api.fixer.io/2016-%s-15?symbols=%s,%s", mm, from, to))
		if err != nil {
			return results
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return results
		}
		var t Fixer
		_ = json.Unmarshal(body, &t)

		// for simplicity, I just store calculated values (to_rate / from_rate) as dataset;
		// if we want more accuracy, we can use 2 different datasets for storing raw data of both 'from' and 'to' rates
		results = append(results, []float64{float64(i), t.Rates.(map[string]interface{})[to].(float64) / t.Rates.(map[string]interface{})[from].(float64)})
	}
	// we can use Redis for caching the results for future use
	return results
}

// Calculate root mean squared error
func rmseMetric(actual, predicted []float64) float64 {
	sum_error := 0.0
	var prediction_error, mean_error float64
	for i := 0; i < len(actual); i++ {
		prediction_error = predicted[i] - actual[i]
		sum_error += math.Pow(prediction_error, 2)
		mean_error = sum_error / float64(len(actual))
	}
	return math.Sqrt(mean_error)
}

// Evaluate regression algorithm on training dataset
func evaluateAlgorithm(dataset [][]float64, algorithm SimpleLinearRegression) float64 {
	var testSet [][]float64
	for _, row := range dataset {
		row_copy := row[:len(row)-1]
		testSet = append(testSet, row_copy)
	}
	predicted := algorithm(dataset, testSet)
	fmt.Println(predicted)
	var actual []float64
	for _, row := range dataset {
		actual = append(actual, row[len(row)-1])
	}
	rmse := rmseMetric(actual, predicted)
	return rmse
}

func sum(a []float64) (sum float64) {
	for _, v := range a {
		sum += v
	}
	return
}

// Calculate the mean value of a list of numbers
func mean(values []float64) float64 {
	return sum(values) / float64(len(values))
}

// Calculate covariance between x and y
func covariance(x []float64, mean_x float64, y []float64, mean_y float64) float64 {
	covar := 0.0
	for i := 0; i < len(x); i++ {
		covar += (x[i] - mean_x) * (y[i] - mean_y)
	}
	return covar
}

// Calculate the variance of a list of numbers
func variance(values []float64, mean float64) float64 {
	sum := 0.0
	for _, x := range values {
		sum += math.Pow(x-mean, 2)
	}
	return sum
}

// Calculate coefficients
func coefficients(dataset [][]float64) []float64 {
	var x, y []float64
	for _, row := range dataset {
		x = append(x, row[0])
		y = append(y, row[1])
	}
	x_mean, y_mean := mean(x), mean(y)
	b1 := covariance(x, x_mean, y, y_mean) / variance(x, x_mean)
	b0 := y_mean - b1 * x_mean
	return []float64{b0, b1}
}

type SimpleLinearRegression func([][]float64, [][]float64) []float64

// Simple linear regression algorithm
func simpleLinearRegression(train, test [][]float64) []float64 {
	var predictions []float64
	b := coefficients(train)
	b0 := b[0]
	b1 := b[1]
	for _, row := range test {
		yhat := b0 + b1 * row[0]
		predictions = append(predictions, yhat)
	}
	return predictions
}

func exchangePredict(dataset [][]float64, algorithm SimpleLinearRegression) float64 {
	var testSet [][]float64
	if (len(dataset) < 12) {
		log.Fatal(errNotEnoughData)
	}
	for _, row := range dataset {
		row_copy := row[:len(row)-1]
		testSet = append(testSet, row_copy)
	}
	predictions := algorithm(dataset, testSet)
	return predictions[len(predictions)-1]
}

func main() {
	if (len(os.Args) < 3) {
		log.Fatal(errNotEnoughArgs)
	}
	from := os.Args[1]
	to := os.Args[2]
	dataSet := loadDataSet(from, to)
	result := exchangePredict(dataSet, simpleLinearRegression)
	fmt.Println(fmt.Sprintf("The predicted currency exchange from %s to %s for 15/1/2017 is %f", from, to, result))
}
