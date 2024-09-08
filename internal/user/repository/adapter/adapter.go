package adapter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	mgo "github.com/core-go/mongo"
	"go-service/internal/user/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewUserAdapter(db *mongo.Database) *UserAdapter {
	bsonMap := mgo.MakeBsonMap(reflect.TypeOf(model.User{}))
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
}

func (r *UserAdapter) All(ctx context.Context) ([]model.User, error) {
	filter := bson.M{}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var users []model.User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserAdapter) Load(ctx context.Context, id string) (*model.User, error) {
	filter := bson.M{"_id": id}
	res := r.Collection.FindOne(ctx, filter)
	if res.Err() != nil {
		if strings.Contains(fmt.Sprint(res.Err()), "mongo: no documents in result") {
			return nil, nil
		} else {
			return nil, res.Err()
		}
	}
	var user model.User
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserAdapter) Create(ctx context.Context, user *model.User) (int64, error) {
	_, err := r.Collection.InsertOne(ctx, user)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key error collection:") {
			if strings.Contains(errMsg, "dup key: { _id: ") {
				return 0, err
			} else {
				return -1, err
			}
		}
		return 0, err
	}
	return 1, nil
}

func (r *UserAdapter) Update(ctx context.Context, user *model.User) (int64, error) {
	filter := bson.M{"_id": user.Id}
	update := bson.M{"$set": user}
	res, err := r.Collection.UpdateOne(ctx, filter, update)
	if res != nil {
		return res.MatchedCount, err
	} else {
		return 0, err
	}
}

func (r *UserAdapter) Patch(ctx context.Context, user map[string]interface{}) (int64, error) {
	id, ok := user["id"]
	if !ok {
		return -1, errors.New("id must be in map[string]interface{} for patch")
	}
	bsonUser := mgo.MapToBson(user, r.Map)
	return mgo.PatchOne(ctx, r.Collection, id, bsonUser)
}

func (r *UserAdapter) Delete(ctx context.Context, id string) (int64, error) {
	filter := bson.M{"_id": id}
	res, err := r.Collection.DeleteOne(ctx, filter)
	if res == nil || err != nil {
		return 0, err
	}
	return res.DeletedCount, err
}
