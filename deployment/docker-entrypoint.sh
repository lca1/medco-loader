#!/usr/bin/env bash
set -Eeuo pipefail

# wait for postgres to be available
export PGPASSWORD="$GA_DB_PASSWORD"
export PSQL_PARAMS="-v ON_ERROR_STOP=1 -h ${GA_DB_HOST} -p ${GA_DB_PORT} -U ${GA_DB_USER}"
until psql $PSQL_PARAMS -d ${GA_DB_NAME} -c '\q'; do
  >&2 echo "Waiting for postgresql..."
  sleep 1
done

QUERY_RESULT=$(psql ${PSQL_PARAMS} -d ${GA_DB_NAME} -X -A -t -c "SELECT count(*) FROM pg_catalog.pg_tables WHERE tablename='genomic_annotations';")

if [[ "$QUERY_RESULT" -eq "0" ]]; then

  SLEEP_TIME=60
  COUNTER=0

  until medco-loader v0 --ont_clinical /test_data/genomic/tcga_cbio/8_clinical_data.csv  --sen /test_data/genomic/sensitive.txt  \
                --ont_genomic /test_data/genomic/tcga_cbio/8_mutation_data.csv  --clinical /test_data/genomic/tcga_cbio/8_clinical_data.csv  \
                --genomic /test_data/genomic/tcga_cbio/8_mutation_data.csv  --output /test_data/ --gaTestData --gaTruncate; do
    >&2 echo "Trying to load test data... ("$COUNTER")"
    COUNTER=$(($COUNTER+1))
    sleep $(($SLEEP_TIME))
  done
  >&2 echo "Test data loaded ("$COUNTER")"

else

>&2 echo "Test data already loaded"

fi