curl -X POST -H "Content-Type: application/json" -d '{
  "package_id": "PKG-2a146e6a-be08-455e-ba39-4ba28fd37e07",
  "weight": 11,
  "from": "Italy",
  "to": "Turki",
  "address": "Baking street",
  "status": "PROCESSED",
  "cost": 1232.66,
  "estimated_hours": 2,
  "currency": "RUB"
}' http://localhost:8333/packages

curl -X POST -H "Content-Type: application/json" -d '{
  "package_id": "PKG-2a146e6a-be08-455e-ba39-4ba28fd37e08",
  "weight": 5,
  "from": "France",
  "to": "Tokio",
  "address": "To the city",
  "status": "PROCESSED",
  "cost": 52312.66,
  "estimated_hours": 8,
  "currency": "RUB"
}' http://localhost:8333/packages

curl -X POST -H "Content-Type: application/json" -d '{
  "package_id": "PKG-2a146e6a-be08-455e-ba39-4ba28fd37e09",
  "weight": 17,
  "from": "Albania",
  "to": "Moscow",
  "address": "Pushkin street",
  "status": "PROCESSED",
  "cost": 5312.66,
  "estimated_hours": 7,
  "currency": "RUB"
}' http://localhost:8333/packages

curl -X POST -H "Content-Type: application/json" -d '{
  "package_id": "PKG-2a146e6a-be08-455e-ba39-4ba28fd37e10",
  "weight": 23,
  "from": "Brazil",
  "to": "Japan",
  "address": "Katarigi",
  "status": "PROCESSED",
  "cost": 2281.13,
  "estimated_hours": 14,
  "currency": "RUB"
}' http://localhost:8333/packages