package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BodyResponse is the returned result
type BodyResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

// BodyInput is the input
type BodyInput struct {
	Doc       map[string]interface{} `json:"docs"`
	URI       string                 `json:"uri"`
	Namespace string                 `json:"namespace"`
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	doc := []byte(request.Body)

	var bodyInput BodyInput
	err := json.Unmarshal(doc, &bodyInput)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not unmarshal JSON string: [%s]", err.Error()))
		return &events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	fmt.Printf("%v\n", bodyInput)
	fmt.Println("Connecting ...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(bodyInput.URI))
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to connect: [%s]", err.Error()))
		return &events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}
	parts := strings.Split(bodyInput.Namespace, ".")
	collection := client.Database(parts[0]).Collection(parts[1])

	results := []string{}
	insertResult, _ := collection.InsertOne(context.TODO(), bodyInput.Doc)
	fmt.Println(insertResult.InsertedID)
	results = append(results, fmt.Sprintf("%v", insertResult.InsertedID))

	fmt.Println(results)
	stringResult := fmt.Sprintf("%v", results)
	response := BodyResponse{Output: stringResult}
	bodyresponse, err := json.Marshal(&response)
	if err != nil {
		return &events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "application/json"},
		Body:            string(bodyresponse),
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(handler)
}
