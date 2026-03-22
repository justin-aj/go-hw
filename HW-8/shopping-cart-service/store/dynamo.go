package store

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"cart-service/model"
)

// DynamoDB implements Store using Amazon DynamoDB.
//
// Single-table design:
//   PK  cartId (S)     — random int stored as string
//   customerId (N)
//   status     (S)     — "active" | "checked_out"
//   items      (L)     — list of {productId (N), quantity (N)} maps

type DynamoDB struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDB loads AWS config from the environment (IAM role on Fargate)
// and returns a ready-to-use DynamoDB store.
func NewDynamoDB(ctx context.Context, region, tableName string) (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("dynamo config: %w", err)
	}
	return &DynamoDB{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

func (d *DynamoDB) CreateCart(ctx context.Context, customerID int) (int, error) {
	cartID := rand.IntN(1_000_000_000) + 1
	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item: map[string]types.AttributeValue{
			"cartId":     &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", cartID)},
			"customerId": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", customerID)},
			"status":     &types.AttributeValueMemberS{Value: "active"},
		},
	})
	if err != nil {
		return 0, err
	}
	return cartID, nil
}

func (d *DynamoDB) GetCart(ctx context.Context, cartID int) (*model.Cart, []model.CartItem, error) {
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"cartId": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", cartID)},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if out.Item == nil {
		return nil, nil, &ErrNotFound{CartID: cartID}
	}

	cart := &model.Cart{ID: cartID}
	if v, ok := out.Item["customerId"].(*types.AttributeValueMemberN); ok {
		cart.CustomerID, _ = strconv.Atoi(v.Value)
	}
	if v, ok := out.Item["status"].(*types.AttributeValueMemberS); ok {
		cart.Status = v.Value
	}

	var items []model.CartItem
	if rawList, ok := out.Item["items"].(*types.AttributeValueMemberL); ok {
		for _, entry := range rawList.Value {
			m, ok := entry.(*types.AttributeValueMemberM)
			if !ok {
				continue
			}
			var item model.CartItem
			if v, ok := m.Value["productId"].(*types.AttributeValueMemberN); ok {
				item.ProductID, _ = strconv.Atoi(v.Value)
			}
			if v, ok := m.Value["quantity"].(*types.AttributeValueMemberN); ok {
				item.Quantity, _ = strconv.Atoi(v.Value)
			}
			items = append(items, item)
		}
	}
	return cart, items, nil
}

func (d *DynamoDB) AddItem(ctx context.Context, cartID, productID, quantity int) error {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"cartId": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", cartID)},
		},
		ConditionExpression:      aws.String("attribute_exists(cartId)"),
		UpdateExpression:         aws.String("SET #it = list_append(if_not_exists(#it, :empty), :item)"),
		ExpressionAttributeNames: map[string]string{"#it": "items"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":item": &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
					"productId": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", productID)},
					"quantity":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", quantity)},
				}},
			}},
			":empty": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return &ErrNotFound{CartID: cartID}
		}
		return err
	}
	return nil
}

func (d *DynamoDB) Checkout(ctx context.Context, cartID int) (int, error) {
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"cartId": &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", cartID)},
		},
		ConditionExpression:      aws.String("attribute_exists(cartId) AND #s = :active"),
		UpdateExpression:         aws.String("SET #s = :checked"),
		ExpressionAttributeNames: map[string]string{"#s": "status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":active":  &types.AttributeValueMemberS{Value: "active"},
			":checked": &types.AttributeValueMemberS{Value: "checked_out"},
		},
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return 0, &ErrNotFound{CartID: cartID}
		}
		return 0, err
	}
	return rand.IntN(1_000_000) + 1, nil
}

func (d *DynamoDB) Close() {}
