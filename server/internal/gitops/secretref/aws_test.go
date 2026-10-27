package secretref

import (
	"errors"
	"net/http"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func TestClassifyAWSError(t *testing.T) {
	response := func(status int, code string) error {
		return &awshttp.ResponseError{ResponseError: &smithyhttp.ResponseError{
			Response: &smithyhttp.Response{Response: &http.Response{StatusCode: status}},
			Err:      &smithy.GenericAPIError{Code: code},
		}}
	}

	testCases := []struct {
		name             string
		err              error
		wantUnresolvable bool
	}{
		{name: "access denied", err: response(400, "AccessDeniedException"), wantUnresolvable: true},
		{name: "secret not found", err: response(400, "ResourceNotFoundException"), wantUnresolvable: true},
		{name: "throttling is retried", err: response(400, "ThrottlingException")},
		{name: "429 is retried", err: response(429, "TooManyRequests")},
		{name: "5xx is retried", err: response(503, "ServiceUnavailable")},
		{name: "network failure (status 0) is retried", err: response(0, "")},
		{name: "error without a response is retried", err: errors.New("dial tcp: no such host")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyAWSError(tc.err)
			if errors.Is(got, ErrUnresolvable) != tc.wantUnresolvable {
				t.Fatalf("errors.Is(got, ErrUnresolvable) = %v, want %v", !tc.wantUnresolvable, tc.wantUnresolvable)
			}
		})
	}
}
