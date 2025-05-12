#!/bin/bash
echo ">>> Importing countries.json into 'logistics.countries'..."
mongoimport --db logistics \
            --collection countries \
            --file /docker-entrypoint-initdb.d/countries.json \
            --jsonArray
echo ">>> Import finished."
