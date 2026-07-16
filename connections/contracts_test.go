package connections

import "testing"

func TestContractsExist(t *testing.T) {
	var _ Database
	var _ Redis
	var _ MongoDB
	var _ S3Client
	var _ GPubSub
	var _ SFTP
	var _ Telemetry
}

func TestDefaultFactorySatisfiesFactory(t *testing.T) {
	var _ Factory = DefaultFactory{}
}
