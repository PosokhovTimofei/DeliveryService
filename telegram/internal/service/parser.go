package service

import (
	"fmt"
	"strings"

	databasepb "github.com/maksroxx/DeliveryService/proto/database"
)

func formatPackageList(pkgs []*databasepb.Package) string {
	if len(pkgs) == 0 {
		return "📭 У вас нет активных посылок."
	}

	var sb strings.Builder
	sb.WriteString("📦 Ваши посылки:\n\n")

	for _, pkg := range pkgs {
		sb.WriteString(fmt.Sprintf("🔹 Заказ: %s\n", pkg.PackageId))
		sb.WriteString(fmt.Sprintf("📍 Откуда: %s → Куда: %s\n", pkg.From, pkg.To))
		sb.WriteString(fmt.Sprintf("📦 Статус: %s\n", pkg.Status))
		sb.WriteString(fmt.Sprintf("💰 %s\n", pkg.PaymentStatus))
		sb.WriteString(fmt.Sprintf("💵 Стоимость: %.2f %s\n", pkg.Cost, pkg.Currency))
		sb.WriteString("─────────────────\n")
	}

	return sb.String()
}
