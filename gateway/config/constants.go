package config

import "os"

var AggregatorService = os.Getenv("AGG_SERVICE_ENDPOINT")
