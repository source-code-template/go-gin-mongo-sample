package adapter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mgo "github.com/core-go/mongo"

	"go-service/internal/user/model"
)

func NewUserAdapter(db *mongo.Database, buildQuery func(*model.UserFilter) (bson.D, bson.M)) *UserAdapter {
	bsonMap := mgo.MakeBsonMap(reflect.TypeOf(model.User{}))
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap, BuildQuery: buildQuery}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
	BuildQuery func(*model.UserFilter) (bson.D, bson.M)
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
		if strings.Compare(fmt.Sprint(res.Err()), "mongo: no documents in result") == 0 {
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

func (r *UserAdapter) Search(ctx context.Context, filter *model.UserFilter, limit int64, offset int64) ([]model.User, int64, error) {
	query, fields := r.BuildQuery(filter)
	var users []model.User
	total, err := r.Collection.CountDocuments(ctx, query)
	if err != nil || total == 0 {
		return users, total, err
	}
	opts := options.Find()
	if len(filter.Sort) > 0 {
		opts.SetSort(mgo.BuildSort(filter.Sort, reflect.TypeOf(model.UserFilter{})))
	}
	opts.SetSkip(offset)
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if fields != nil {
		opts.Projection = fields
	}
	cursor, err := r.Collection.Find(ctx, query, opts)
	if err != nil {
		return users, total, err
	}
	err = cursor.All(ctx, &users)
	return users, total, err
}
