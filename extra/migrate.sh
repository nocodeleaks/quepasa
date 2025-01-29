#!/bin/sh
####################################
##	SUFFICIT SHELL SCRIPT
##	All Rights Reserved (2025) SUFFICIT SOLUCOES EM TECNOLOGIA DA INFORMACAO.
##	This script will help you to migrate from SQLite to Postgres the whatsmeow database
##	Version 1.0.0
##	Usage: migrate.sh {postgres_user} {postgres_db} {postgres_password}   
####################################

set -e
DUMPFILE=/mnt/migrate.dump
SQLITEFILE=/opt/quepasa-source/src/whatsmeow.sqlite

# Creating dump file, only inserts
sqlite3 ${SQLITEFILE} .dump | grep 'INSERT' | grep -v 'replace(' > ${DUMPFILE}

# Adjusting blobs fields
sed -e s/\,X\'/\,\'\\\\x/g -i ${DUMPFILE}

# Adjusting Booleans
sed -e '/whatsmeow_pre_keys/ s/\,1)/\,true)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_pre_keys/ s/\,0)/\,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,1,0)/\,true,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,0,0)/\,false,false)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,0,1)/\,false,true)/g' -i ${DUMPFILE}
sed -e '/whatsmeow_chat_settings/ s/\,1,1)/\,true,true)/g' -i ${DUMPFILE}

# Adjusting sqlite to postgres fields
sed -i -e 's/INTEGER PRIMARY KEY AUTOINCREMENT/SERIAL PRIMARY KEY/g;s/PRAGMA foreign_keys=OFF;//;s/unsigned big int/BIGINT/g;s/UNSIGNED BIG INT/BIGINT/g;s/BIG INT/BIGINT/g;s/UNSIGNED INT(10)/BIGINT/g;s/BOOLEAN/SMALLINT/g;s/boolean/SMALLINT/g;s/UNSIGNED BIG INT/INTEGER/g;s/INT(3)/INT2/g;s/DATETIME/TIMESTAMP/g' ${DUMPFILE}

# Truncating tables
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

# Prepend truncate commands
echo "${TRUNCATETEXT}$(cat ${DUMPFILE})" > ${DUMPFILE}

#echo "SET session_replication_role TO 'replica';\n$(cat ${DUMPFILE})" > ${DUMPFILE}
echo "SET CONSTRAINTS ALL DEFERRED;\n$(cat ${DUMPFILE})" > ${DUMPFILE}
echo "BEGIN;\n$(cat ${DUMPFILE})" > ${DUMPFILE}

echo "COMMIT;" >> ${DUMPFILE}

psql postgresql://${1}:${3}@127.0.0.1/${2} < ${DUMPFILE} > ${DUMPFILE}.transaction 2>&1
