#!/bin/bash

BASE_URL="http://localhost:8121"

get_tariffs() {
    echo ">>> Получение тарифов..."
    curl -X GET "$BASE_URL/tariffs"
    echo -e "\n"
}

calculate_by_tariff() {
    echo ">>> Расчет по тарифу..."
    curl -X POST "$BASE_URL/calculate-by-tariff" \
        -H "Content-Type: application/json" \
        -d '{
            "weight": 2.5,
            "from": "Russia",
            "to": "Germany",
            "address": "Lenina 10",
            "length": 30,
            "width": 20,
            "height": 15,
            "tariff_code": "EXPRESS"
        }'
    echo "\n" 
}

calculate() {
    echo ">>> Общий расчет..."
    curl -X POST "$BASE_URL/calculate" \
        -H "Content-Type: application/json" \
        -d '{
            "weight": 2.5,
            "from": "Russia",
            "to": "Germany",
            "address": "Lenina 10",
            "length": 30,
            "width": 20,
            "height": 15
        }'
    echo "\n"
}

get_tariffs
calculate_by_tariff
calculate