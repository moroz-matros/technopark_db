#!/bin/bash

sudo su postgres -c "psql -c \"CREATE user forum_postgre WITH PASSWORD 'forum_postgre'; \""
sudo su postgres -c "psql -c \"create database forum owner forum_postgre; \""
sudo su postgres -c "psql -c \"grant all privileges on database forum to forum_postgre; \""
PGPASSWORD=forum_postgre psql -U forum_postgre -h localhost -d forum -f scripts/start.sql
