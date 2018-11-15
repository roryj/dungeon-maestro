GOOS=linux GOARCH=amd64 go build -o main src/go.roryj.dnd/main.go

zip main.zip main

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
