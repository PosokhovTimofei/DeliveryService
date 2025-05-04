package transport

import (
	"regexp"
	"strings"

	"github.com/maksroxx/DeliveryService/calculator/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateAddress(pkg models.Package) error {
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		err := status.Error(codes.InvalidArgument, "Missing required address fields")
		return err
	}

	if isOnlyDigits(pkg.Address) || isOnlyDigits(pkg.From) || isOnlyDigits(pkg.To) {
		err := status.Error(codes.InvalidArgument, "Address cannot consist only of digits")
		return err
	}

	return nil
}

func isOnlyDigits(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(strings.TrimSpace(s))
}
