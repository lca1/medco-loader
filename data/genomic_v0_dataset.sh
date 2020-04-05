#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_FOLDER="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_URL="https://github.com/ldsec/projects-data/blob/master/medco/datasets"

# genomic (v0) dataset
mkdir -p ${SCRIPT_FOLDER}/genomic/tcga_cbio/
wget -O ${SCRIPT_FOLDER}/genomic/tcga_cbio/8_mutation_data.csv ${REPO_URL}/genomic/tcga_cbio/8_mutation_data.csv?raw=true
wget -O ${SCRIPT_FOLDER}/genomic/tcga_cbio/8_clinical_data.csv ${REPO_URL}/genomic/tcga_cbio/8_clinical_data.csv?raw=true
wget -O ${SCRIPT_FOLDER}/genomic/sensitive.txt ${REPO_URL}/genomic/sensitive.txt?raw=true
wget -O ${SCRIPT_FOLDER}/genomic/group.toml ${REPO_URL}/genomic/group.toml?raw=true