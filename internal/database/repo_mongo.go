package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository[T Entity] struct {
	client     *mongo.Client
	collection *mongo.Collection
	uri        string
	dbName     string
	collName   string
}

func NewMongoRepository[T Entity](uri, dbName, collName string) *MongoRepository[T] {
	return &MongoRepository[T]{
		uri:      uri,
		dbName:   dbName,
		collName: collName,
	}
}

func (r *MongoRepository[T]) Connect(ctx context.Context) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.uri))
	if err != nil {
		return fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ошибка ping MongoDB: %w", err)
	}

	r.client = client
	r.collection = client.Database(r.dbName).Collection(r.collName)

	return nil
}

func (r *MongoRepository[T]) Close(ctx context.Context) error {
	if r.client != nil {
		return r.client.Disconnect(ctx)
	}
	return nil
}

func (r *MongoRepository[T]) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

func (r *MongoRepository[T]) Create(ctx context.Context, entity T) error {
	if entity.GetID() == "" {
		entity.SetID(primitive.NewObjectID().Hex())
	}

	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *MongoRepository[T]) Get(ctx context.Context, id string) (T, error) {
	var entity T

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return entity, fmt.Errorf("неверный ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	err = r.collection.FindOne(ctx, filter).Decode(&entity)
	if err == mongo.ErrNoDocuments {
		return entity, nil
	}

	return entity, err
}

func (r *MongoRepository[T]) Update(ctx context.Context, entity T) error {
	objID, err := primitive.ObjectIDFromHex(entity.GetID())
	if err != nil {
		return fmt.Errorf("неверный ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": entity}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("запись не найдена")
	}

	return nil
}

func (r *MongoRepository[T]) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("неверный ID: %w", err)
	}

	filter := bson.M{"_id": objID}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("запись не найдена")
	}

	return nil
}

func (r *MongoRepository[T]) Find(ctx context.Context, filter Filter, opts *Options) ([]T, error) {
	mongoFilter := r.convertFilter(filter)

	findOpts := options.Find()
	if opts != nil {
		if opts.Limit > 0 {
			findOpts.SetLimit(opts.Limit)
		}
		if opts.Offset > 0 {
			findOpts.SetSkip(opts.Offset)
		}
		if opts.Sort != nil {
			sortDoc := bson.D{}
			for field, order := range opts.Sort {
				sortDoc = append(sortDoc, bson.E{Key: field, Value: order})
			}
			findOpts.SetSort(sortDoc)
		}
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoRepository[T]) FindOne(ctx context.Context, filter Filter) (T, error) {
	var entity T

	mongoFilter := r.convertFilter(filter)
	err := r.collection.FindOne(ctx, mongoFilter).Decode(&entity)
	if err == mongo.ErrNoDocuments {
		return entity, nil
	}

	return entity, err
}

func (r *MongoRepository[T]) Count(ctx context.Context, filter Filter) (int64, error) {
	mongoFilter := r.convertFilter(filter)
	return r.collection.CountDocuments(ctx, mongoFilter)
}

func (r *MongoRepository[T]) CreateMany(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	docs := make([]interface{}, len(entities))
	for i, entity := range entities {
		if entity.GetID() == "" {
			entity.SetID(primitive.NewObjectID().Hex())
		}
		docs[i] = entity
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *MongoRepository[T]) UpdateMany(ctx context.Context, filter Filter, update map[string]interface{}) error {
	mongoFilter := r.convertFilter(filter)
	mongoUpdate := bson.M{"$set": update}

	_, err := r.collection.UpdateMany(ctx, mongoFilter, mongoUpdate)
	return err
}

func (r *MongoRepository[T]) DeleteMany(ctx context.Context, filter Filter) error {
	mongoFilter := r.convertFilter(filter)
	_, err := r.collection.DeleteMany(ctx, mongoFilter)
	return err
}

func (r *MongoRepository[T]) convertFilter(filter Filter) bson.M {
	if filter == nil {
		return bson.M{}
	}

	result := bson.M{}
	for k, v := range filter {
		switch val := v.(type) {
		case map[string]interface{}:
			// Операторы сравнения
			ops := bson.M{}
			for op, opVal := range val {
				ops["$"+op] = opVal // gt -> $gt, lt -> $lt, etc
			}
			result[k] = ops
		default:
			result[k] = v
		}
	}
	return result
}
