#!/usr/bin/env bash
set -Eeuo pipefail

### script to download the MedCo example datasets

# download genomic (v0) dataset
bash ./genomic_v0_dataset.sh

# download i2b2 demo (v1) dataset
bash ./i2b2_v1_dataset.sh
