#!/bin/bash

CLIENT="./bin/client"

print_separator() {
    echo "----------------------------------------------------"
}

echo "Собираем проект..."
make client

# Тест 1: Создание посылки
print_separator
echo "Тест 1: Создание посылки"
$CLIENT create 5.2 Москва "Санкт-Петербург" "ул. Ленина, 1"

# Тест 2: Проверка статуса
print_separator
echo "Тест 2: Проверка статуса посылки"
$CLIENT status $ID1

# Тест 3: Неверный вес
print_separator
echo "Тест 3: Попытка создать посылку с отрицательным весом"
$CLIENT create -5.0 Москва Казань "ул. Пушкина, 15"

# Тест 4: Неполный адрес
print_separator
echo "Тест 4: Создание посылки без указания адреса"
$CLIENT create 8.1 Москва "Нижний Новгород"

# Тест 5: Проверка несуществующего ID
print_separator
echo "Тест 5: Проверка ID"
$CLIENT status "PKG-2a146e6a-be08-455e-ba39-4ba28fd37e07"

print_separator
echo "Тестирование завершено"