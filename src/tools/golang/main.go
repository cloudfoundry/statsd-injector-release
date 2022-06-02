package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

func main() {
	port := "8125"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	c := &statsd.ClientConfig{
		Address: fmt.Sprintf("127.0.0.1:%s", port),
		Prefix:  "testNamespace",
	}
	client, err := statsd.NewClientWithConfig(c)
	if err != nil {
		fmt.Printf("Error connecting to statsd server: %v\n", err.Error())
		return
	}
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		inputs := strings.Split(line, " ")
		if len(inputs) < 3 {
			fmt.Printf("Wrong number of inputs, 3 needed at least\n")
			continue
		}
		statsdType := inputs[0]
		name := inputs[1]

		value, _ := strconv.ParseInt(inputs[2], 10, 0)
		var sampleRate float32 = 1.0
		if len(inputs) == 4 {
			rate, _ := strconv.ParseFloat(inputs[3], 32)
			sampleRate = float32(rate)
		}

		switch statsdType {
		case "count":
			err = client.Inc(name, value, sampleRate)
			fmt.Println("Failed to increment counter:", err)
		case "gauge":
			err = client.Gauge(name, value, sampleRate)
			fmt.Println("Failed to submit/update gauge:", err)
		case "gaugedelta":
			err = client.GaugeDelta(name, value, sampleRate)
			fmt.Println("Failed to submit delta:", err)
		case "timing":
			err = client.Timing(name, value, sampleRate)
			fmt.Println("Failed to submit timing stat:", err)
		default:
			fmt.Printf("Unsupported operation: %s\n", statsdType)
		}
	}
}
