#!/bin/bash

go build -o bookings ./cmd/web/*.go && \
./bookings \
-production=false \
-cache=false \
-dbhost=127.0.0.1 \
-dbport=5432 \
-dbname=bookings \
-dbuser=postgres \
-dbpassword=password
