package transport_test

import (
	"testing"

	"github.com/maksroxx/DeliveryService/calculator/internal/transport"
	"github.com/maksroxx/DeliveryService/calculator/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name     string
		pkg      models.Package
		expected codes.Code
	}{
		{
			name: "Valid address",
			pkg: models.Package{
				From:    "Sender",
				To:      "Receiver",
				Address: "123 Main St",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.OK,
		},
		{
			name: "Missing From",
			pkg: models.Package{
				From:    "",
				To:      "Receiver",
				Address: "123 Main St",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "Missing To",
			pkg: models.Package{
				From:    "Sender",
				To:      "",
				Address: "123 Main St",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "Missing Address",
			pkg: models.Package{
				From:    "Sender",
				To:      "Receiver",
				Address: "",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "Address consists only of digits",
			pkg: models.Package{
				From:    "Sender",
				To:      "Receiver",
				Address: "123456",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "From consists only of digits",
			pkg: models.Package{
				From:    "123456",
				To:      "Receiver",
				Address: "123 Main St",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
		{
			name: "To consists only of digits",
			pkg: models.Package{
				From:    "Sender",
				To:      "123456",
				Address: "123 Main St",
				Length:  2,
				Height:  2,
				Width:   2,
			},
			expected: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transport.Validate(tt.pkg)
			if tt.expected == codes.OK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, status.Code(err))
			}
		})
	}
}
