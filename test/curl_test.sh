#!/bin/sh

echo $1

for ((i = 1; i <= $1; i++)); do
    curl -s -X POST -d '{"mailbox":"home.foosa"}' "http://localhost:3112/get" &
done
