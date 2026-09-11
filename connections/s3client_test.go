package connections

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	helperconfig "github.com/radian-solusi/microservice-helpers/config"
)

func TestGCSProviderDefaultsToXMLEndpoint(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{
		Provider:        helperconfig.S3ProviderGCS,
		BucketName:      "my-bucket",
		AccessKeyID:     "GOOGACCESSKEY",
		SecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if c.IsLocalStorage() {
		t.Fatal("GCS client must not report as local storage")
	}
	client := c.Client()
	if client == nil {
		t.Fatal("GCS provider must use the AWS SDK *s3.Client (HMAC via S3 XML API)")
	}
	opts := client.Options()
	if opts.BaseEndpoint == nil || *opts.BaseEndpoint != gcsEndpoint {
		t.Fatalf("BaseEndpoint = %v, want %s", opts.BaseEndpoint, gcsEndpoint)
	}
	if !opts.UsePathStyle {
		t.Fatal("expected path-style addressing for GCS")
	}
	if opts.Region != "auto" {
		t.Fatalf("Region = %q, want auto", opts.Region)
	}
	if opts.RequestChecksumCalculation != aws.RequestChecksumCalculationWhenRequired {
		t.Fatalf("RequestChecksumCalculation = %v, want WhenRequired", opts.RequestChecksumCalculation)
	}
	if opts.ResponseChecksumValidation != aws.ResponseChecksumValidationWhenRequired {
		t.Fatalf("ResponseChecksumValidation = %v, want WhenRequired", opts.ResponseChecksumValidation)
	}
}

func TestGCSProviderHonorsExplicitEndpointAndRegion(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{
		Provider:        helperconfig.S3ProviderGCS,
		BucketName:      "my-bucket",
		Endpoint:        "http://127.0.0.1:4443", // e.g. fake-gcs-server
		Region:          "us-east-1",
		AccessKeyID:     "GOOGACCESSKEY",
		SecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	opts := c.Client().Options()
	if opts.BaseEndpoint == nil || *opts.BaseEndpoint != "http://127.0.0.1:4443" {
		t.Fatalf("BaseEndpoint = %v, want explicit override", opts.BaseEndpoint)
	}
	if opts.Region != "us-east-1" {
		t.Fatalf("Region = %q, want explicit override", opts.Region)
	}
}

func TestGCSGetFileExtension(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{
		Provider:        helperconfig.S3ProviderGCS,
		BucketName:      "my-bucket",
		AccessKeyID:     "GOOGACCESSKEY",
		SecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if got := c.GetFileExtension("docs/report.PDF"); got != ".PDF" {
		t.Fatalf("extension: got %q", got)
	}
}

func TestAWSProviderUnaffectedByGCSDefaults(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{
		Provider:        helperconfig.S3ProviderAWS,
		BucketName:      "my-bucket",
		Region:          "ap-southeast-1",
		AccessKeyID:     "AKIA",
		SecretAccessKey: "secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	opts := c.Client().Options()
	if opts.BaseEndpoint != nil {
		t.Fatalf("BaseEndpoint = %v, want nil for default AWS endpoint resolution", opts.BaseEndpoint)
	}
	if opts.Region != "ap-southeast-1" {
		t.Fatalf("Region = %q, want ap-southeast-1", opts.Region)
	}
	if opts.RequestChecksumCalculation == aws.RequestChecksumCalculationWhenRequired {
		t.Fatal("AWS provider must keep SDK default checksum behavior, not the GCS override")
	}
}

func TestMinioProviderUsesPathStyleWithoutGCSDefaults(t *testing.T) {
	c, err := NewS3Client(context.Background(), helperconfig.S3Config{
		Provider:        helperconfig.S3ProviderMinio,
		BucketName:      "my-bucket",
		Endpoint:        "http://127.0.0.1:9000",
		Region:          "us-east-1",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	opts := c.Client().Options()
	if opts.BaseEndpoint == nil || *opts.BaseEndpoint != "http://127.0.0.1:9000" {
		t.Fatalf("BaseEndpoint = %v, want minio endpoint", opts.BaseEndpoint)
	}
	if !opts.UsePathStyle {
		t.Fatal("expected path-style addressing for minio")
	}
	if opts.RequestChecksumCalculation == aws.RequestChecksumCalculationWhenRequired {
		t.Fatal("minio provider must keep SDK default checksum behavior, not the GCS override")
	}
}
