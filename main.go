package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/go-awesome/nft-metadata/config"
	"github.com/gofiber/fiber/v2"
	"io"
	"io/ioutil"
	"net/http"
)

type NFTMetadata struct {
	IpfsHash string                 `json:"ipfs_hash"`
	Metadata map[string]interface{} `json:"metadata"`
}

func main() {
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "NFT Metadata",
		AppName:       "NFT Metadata fetching ",
	})

	app.Get("/tokens", func(c *fiber.Ctx) error {
		metadata := getAllFromDynamodb()
		return c.JSON(metadata)
	})

	app.Get("/tokens/:cid", func(c *fiber.Ctx) error {
		ipfsHash := c.Params("cid")
		metadata := fetchIpfs(ipfsHash)
		return c.JSON(saveToDynamodb(ipfsHash, metadata))
	})

	app.Get("/get-db/:ipfs_hash", func(c *fiber.Ctx) error {
		ipfsHash := c.Params("ipfs_hash")
		metadata := getFromDynamodb(ipfsHash)
		return c.JSON(metadata)
	})

	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err.Error())
	}

}

func fetchIpfs(ipfsHash string) map[string]interface{} {
	url := config.NFT_URL + ipfsHash
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(response.Body)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return result

}

func saveToDynamodb(ipfsHash string, metadata map[string]interface{}) string {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWS_REGION)},
	)
	if err != nil {
		fmt.Println(err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Create item in table
	av, err := dynamodbattribute.MarshalMap(NFTMetadata{
		IpfsHash: ipfsHash,
		Metadata: metadata,
	})
	if err != nil {
		fmt.Println(err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(config.TABLE_NAME),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println(err)
	}
	return "success"

}

func getFromDynamodb(ipfsHash string) map[string]interface{} {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		fmt.Println(err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	ipfsFilter := expression.Name("ipfs_hash").Equal(expression.Value(ipfsHash))

	// Get back the title, year, and rating
	proj := expression.NamesList(expression.Name("ipfs_hash"), expression.Name("metadata"))

	expr, err := expression.NewBuilder().WithFilter(ipfsFilter).WithProjection(proj).Build()
	if err != nil {
		fmt.Println(err)
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(config.TABLE_NAME),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err)
	}

	var nftMetadata NFTMetadata
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &nftMetadata)
	if err != nil {
		fmt.Println(err)
	}
	return nftMetadata.Metadata
}

func getAllFromDynamodb() []map[string]interface{} {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	if err != nil {
		fmt.Println(err)
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	// Create the Expression to fill the input struct with.
	ipfsFilter := expression.Name("ipfs_hash").Equal(expression.Value(""))

	// Get back the title, year, and rating
	proj := expression.NamesList(expression.Name("ipfs_hash"), expression.Name("metadata"))

	expr, err := expression.NewBuilder().WithFilter(ipfsFilter).WithProjection(proj).Build()
	if err != nil {
		fmt.Println(err)
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(config.TABLE_NAME),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println(err)
	}

	var nftMetadata []NFTMetadata
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &nftMetadata)
	if err != nil {
		fmt.Println(err)
	}
	var metadata []map[string]interface{}
	for _, data := range nftMetadata {
		metadata = append(metadata, data.Metadata)
	}
	return metadata
}
