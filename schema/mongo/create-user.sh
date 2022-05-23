#!/usr/bin/env bash
echo "Creating mongo users..."
mongo admin --host localhost -u root -p root --authenticationDatabase admin --eval "db.createUser({user: 'replaceme', pwd: 'replaceme', roles: [{role: 'readWrite', db: 'hen-salonlab'},{role: 'dbAdmin', db: 'non-existing-dbname'}]});"
echo "Mongo users created."
