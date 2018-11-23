#!/bin/bash

# Change to the directory with our code that we plan to work from
cd "$GOPATH/src/github.com/joncalhoun/twg"

echo "==== Releasing gopherswag.com ===="
echo "  Deleting the local binary if it exists (so it isn't uploaded)..."
rm gopherswag.com
echo "  Done!"

echo "  Deleting existing code..."
ssh root@gopherswag.com "rm -rf /root/go/src/github.com/joncalhoun/twg"
echo "  Code deleted successfully!"


echo "  Uploading code..."
ssh root@gopherswag.com "mkdir -p /root/go/src/github.com/joncalhoun/twg/"
# The \ at the end of the line tells bash that our
# command isn't done and wraps to the next line.
rsync -avr --exclude '.git/*' --exclude 'tmp/*' \
  --exclude 'images/*' ./ \
  root@gopherswag.com:/root/go/src/github.com/joncalhoun/twg/
echo "  Code uploaded successfully!"

echo "  Building the code on remote server..."
ssh root@gopherswag.com 'export GOPATH=/root/go; \
  cd $GOPATH/src/github.com/joncalhoun/twg; \
  /usr/local/go/bin/go build -o /root/app/server \
    ./swag/*.go'
echo "  Code built successfully!"

echo "  Moving assets..."
ssh root@gopherswag.com "cd /root/app; \
  cp -R /root/go/src/github.com/joncalhoun/twg/swag/assets ."
echo "  Assets moved successfully!"

echo "  Moving templates..."
ssh root@gopherswag.com "cd /root/app; \
  cp -R /root/go/src/github.com/joncalhoun/twg/swag/templates ."
echo "  Templates moved successfully!"

echo "  Moving Caddyfile..."
ssh root@gopherswag.com "cd /root/app; \
  cp /root/go/src/github.com/joncalhoun/twg/swag/Caddyfile ."
echo "  Caddyfile moved successfully!"

echo "  Moving ENV file..."
ssh root@gopherswag.com "cd /root/app; \
  cp /root/go/src/github.com/joncalhoun/twg/swag/.env ."
echo "  ENV file moved successfully!"


echo "  Moving migration files..."
ssh root@gopherswag.com "cd /root/app; \
  cp -R /root/go/src/github.com/joncalhoun/twg/swag/db/migrations ."
echo "  Migration files moved successfully!"

# echo "  Running migrations..."
# ssh root@gopherswag.com "cd /root/app; \
#   ./migrate.linux-amd64 -source file://migrations -database \"postgres://localhost:5432/swag_prod?user=postgres&password=2fY%23bL83&sslmode=disable&dbname=swag_prod\" up"
# echo "  Migrations done!"

echo "  Restarting the server..."
ssh root@gopherswag.com "sudo service gopherswag.com restart"
echo "  Server restarted successfully!"

echo "  Restarting Caddy server..."
ssh root@gopherswag.com "sudo service caddy restart"
echo "  Caddy restarted successfully!"

echo "==== Done releasing gopherswag.com ===="
