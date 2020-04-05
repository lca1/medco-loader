[![Build Status](https://travis-ci.org/ldsec/medco-loader.svg?branch=master)](https://travis-ci.org/ldsec/medco-loader) 
[![Go Report Card](https://goreportcard.com/badge/github.com/ldsec/medco-loader)](https://goreportcard.com/report/github.com/ldsec/medco-loader) 
[![Coverage Status](https://coveralls.io/repos/github/ldsec/medco-loader/badge.svg?branch=master)](https://coveralls.io/github/ldsec/medco-loader?branch=master)

## medco-loader
*medco-loader* is an ETL tool to encrypt and load data into MedCo.

## Getting started
Run the following commands to download and build the *medco-loader* module.
```shell
git clone https://github.com/ldsec/medco-loader.git
cd medco-loader/deployment/
docker-compose build
``` 

Before using the *medco-loader* you need to have MedCo up and running on your machine. To achieve that you can follow, for example, the [Local Development Deployment guide](https://ldsec.gitbook.io/medco-documentation/developers/local-development-deployment). 

## How to use it
A detailed up-to-date guide on how to use the *medco-loader* is available [here](https://ldsec.gitbook.io/medco-documentation/system-administrators/data-loading).

## Source code organization
- *app*: *medco-loader* command line interface
- *deployment*: docker configuration files
- *loader*: *medco-loader* logic
    - *genomic*: genomic loader logic
    - *i2b2*: i2b2 loader logic
    - *identifiers*: logic managing the identifiers that are meant to be encrypted by [unlynx](https://github.com/ldsec/unlynx) to answer queries

## Useful information
*medco-loader* is part of the MedCo system.

You can find more information about the MedCo project [here](https://medco.epfl.ch/).

For further details, support, and contacts, you can check the [MedCo Technical Documentation](https://ldsec.gitbook.io/medco-documentation/).

## License
*medco-loader* is licensed under a End User Software License Agreement ('EULA') for non-commercial use.
If you need more information, please contact us.
