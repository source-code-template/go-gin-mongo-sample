package query

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	mgo "github.com/core-go/mongo"

	"go-service/internal/user/model"
)

func BuildQuery(filter *model.UserFilter) (bson.D, bson.M) {
	query := bson.D{}
	if len(filter.Id) > 0 {
		query = append(query, bson.E{Key: "_id", Value: filter.Id})
	}
	if filter.DateOfBirth != nil {
		dobFilter := bson.M{}
		if filter.DateOfBirth.Min != nil {
			dobFilter["$gte"] = filter.DateOfBirth.Min
		}
		if filter.DateOfBirth.Max != nil {
			dobFilter["$lte"] = filter.DateOfBirth.Max
		}
		if len(dobFilter) > 0 {
			query = append(query, bson.E{Key: "dateOfBirth", Value: dobFilter})
		}
	}
	if len(filter.Username) > 0 {
		query = append(query, bson.E{Key: "username", Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", filter.Username), Options: "i"}})
	}
	if len(filter.Email) > 0 {
		query = append(query, bson.E{Key: "email", Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", filter.Email), Options: "i"}})
	}
	if len(filter.Phone) > 0 {
		query = append(query, bson.E{Key: "phone", Value: primitive.Regex{Pattern: fmt.Sprintf("\\w*%v\\w*", filter.Phone), Options: "i"}})
	}

	fields := mgo.GetFields(filter.Fields, reflect.TypeOf(model.User{}))
	return query, fields
}
