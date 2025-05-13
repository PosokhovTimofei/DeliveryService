#!/bin/bash
echo ">>> Importing countries.json into 'logistics.countries'..."
mongoimport --db logistics \
            --collection countries \
            --file /docker-entrypoint-initdb.d/countries.json \
            --jsonArray
echo ">>> Import finished."

echo ">>> Importing tariffs.json into 'logistics.tariffs'..."
mongoimport --db logistics \
            --collection tariffs \
            --file /docker-entrypoint-initdb.d/tariffs.json \
            --jsonArray
echo ">>> Tariffs import finished."