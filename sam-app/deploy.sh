#!/usr/bin/env bash

make clean
make build || exit 1

zip maestro.zip maestro

sam package --template-file template.yaml \
    --output-template-file packaged.yaml \
    --s3-bucket dungeon-maestro-repo \
    --profile roryj \
    --region us-west-2

sam deploy \
    --template-file packaged.yaml \
    --stack-name dungeon-maestro-stack \
    --capabilities CAPABILITY_NAMED_IAM \
    --profile roryj \
    --region us-west-2
