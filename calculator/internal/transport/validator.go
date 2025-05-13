package transport

import (
	"regexp"
	"strings"

	"github.com/maksroxx/DeliveryService/calculator/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Validate(pkg models.Package) error {
	if pkg.From == "" || pkg.To == "" || pkg.Address == "" {
		err := status.Error(codes.InvalidArgument, "Missing required address fields")
		return err
	}

	if pkg.Length <= 0 || pkg.Width <= 0 || pkg.Height <= 0 {
		err := status.Error(codes.InvalidArgument, "Invalid parameters")
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
