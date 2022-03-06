package mongodb

import (
	"context"
	"fmt"
	"time"

	rep "github.com/Ataraxxia/godin/report"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	User         string
	Password     string
	DatabaseName string
	URI          string
	//	MockDB        *sql.DB
}

type MongoDBConfiguration struct {
	User         string
	Password     string
	DatabaseName string
	URI          string
}

var (
	ctx context.Context
)

func (d DB) InitializeDatabase() error {
	return nil
}

func (d DB) getDatabaseHandle() (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(fmt.Sprintf("%s", d.URI)))
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (d DB) SaveReport(r rep.Report, t time.Time) error {
	client, err := d.getDatabaseHandle()
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	collection := client.Database(d.DatabaseName).Collection("reports")

	v, _ := r.Value()
	_, err = collection.InsertOne(ctx, v)
	if err != nil {
		return err
	}

	return nil
}
