#!/bin/sh
####################################
##  SUFFICIT SHELL SCRIPT
##  All Rights Reserved (2025) SUFFICIT SOLUCOES EM TECNOLOGIA DA INFORMACAO.
##  This script will help you to migrate from SQLite to Postgres the whatsmeow database
##  Version 1.0.1
##  Usage: migrate.sh {postgres_user} {postgres_db} {postgres_password}   
####################################

set -e
echo "Starting the migration process from SQLite to PostgreSQL..."
DUMPFILE=/mnt/migrate.dump
SQLITEFILE=/opt/quepasa-source/src/whatsmeow.sqlite

echo "Creating the dump file from the SQLite database: ${SQLITEFILE}"
sqlite3 ${SQLITEFILE} .dump | grep 'INSERT' | grep -v 'replace(' > ${DUMPFILE}
echo "Dump file created at ${DUMPFILE}"

echo "Adjusting blob fields in the dump file..."
sed -e s/\,X\'/\,\'\\\\x/g -i ${DUMPFILE}
echo "Blob fields adjusted."

echo "Adjusting boolean fields in the dump file..."
sed -e '/whatsmeow_pre_keys/ s/\,1)/\,true)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_pre_keys/ s/\,0)/\,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,1,0)/\,true,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,0,0)/\,false,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,0,1)/\,false,true)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,1,1)/\,true,true)/g' -i ${DUMPFILE}
echo "Boolean fields adjusted."

echo "Converting SQLite data types to PostgreSQL equivalents..."
sed -i -e 's/INTEGER PRIMARY KEY AUTOINCREMENT/SERIAL PRIMARY KEY/g;s/PRAGMA foreign_keys=OFF;//;s/unsigned big int/BIGINT/g;s/UNSIGNED BIG INT/BIGINT/g;s/BIG INT/BIGINT/g;s/UNSIGNED INT(10)/BIGINT/g;s/BOOLEAN/SMALLINT/g;s/boolean/SMALLINT/g;s/UNSIGNED BIG INT/INTEGER/g;s/INT(3)/INT2/g;s/DATETIME/TIMESTAMP/g' ${DUMPFILE}
echo "Data types conversion complete."

echo "Preparing table truncation commands..."
TRUNCATETEXT=''
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_app_state_mutation_macs RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_app_state_sync_keys RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_app_state_version RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_chat_settings RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_contacts RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_device RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_identity_keys RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_pre_keys RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_privacy_tokens RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_sender_keys RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_sessions RESTART IDENTITY CASCADE;\n"
TRUNCATETEXT="${TRUNCATETEXT}TRUNCATE TABLE whatsmeow_version RESTART IDENTITY CASCADE;\n"
echo "Table truncation commands prepared."

echo "Prepending truncation commands to the dump file..."
echo "${TRUNCATETEXT}$(cat ${DUMPFILE})" > ${DUMPFILE}
echo "Truncation commands prepended."

echo "Setting constraints to DEFERRED..."
echo "SET CONSTRAINTS ALL DEFERRED;\n$(cat ${DUMPFILE})" > ${DUMPFILE}
echo "Constraints set to DEFERRED."

echo "Starting transaction block..."
echo "BEGIN;\n$(cat ${DUMPFILE})" > ${DUMPFILE}
echo "Transaction block started."

echo "Appending COMMIT command at the end of the dump file..."
echo "COMMIT;" >> ${DUMPFILE}
echo "COMMIT command appended."

echo "Executing migration to PostgreSQL..."
psql postgresql://${1}:${3}@127.0.0.1/${2} < ${DUMPFILE} > ${DUMPFILE}.transaction 2>&1
echo "Migration executed. Please check the file ${DUMPFILE}.transaction for transaction details."
