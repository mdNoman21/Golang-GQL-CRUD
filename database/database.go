package database

import (
	"context"
	"gql-yt/graph/model"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString string = "mongodb://localhost:27017"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	return &DB{
		client: client,
	}
}

func (db *DB) GetJob(jobId string) *model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(jobId)

	filter := bson.M{"_id": _id}
	var jobListing model.JobListing
	err = jobCollection.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		log.Fatal(err)
	}
	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListings []*model.JobListing
	cursor, err := jobCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context.TODO(), &jobListings); err != nil {
		panic(err)
	}
	// for cursor.Next(ctx) {
	// 	var job model.JobListing
	// 	err = cursor.Decode(&job)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	jobListings = append(jobListings, &job)
	// }

	return jobListings
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {

	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	inserted, err := jobCollection.InsertOne(ctx, bson.M{
		"title":       jobInfo.Title,
		"description": jobInfo.Description,
		"company":     jobInfo.Company,
		"URL":         jobInfo.URL,
	})
	if err != nil {
		log.Fatal(err)
	}
	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()
	returnJobListing := model.JobListing{
		ID:          insertedID,
		Title:       jobInfo.Title,
		Description: jobInfo.Description,
		Company:     jobInfo.Company,
		URL:         jobInfo.URL,
	}
	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// _id, err := primitive.ObjectIDFromHex(jobId)
	// filter := bson.M{"_id": _id}
	// update := bson.M{
	// 	"$set": bson.M{
	// 		"title":       jobInfo.Title,
	// 		"description": jobInfo.Description,
	// 		"url":         jobInfo.URL,
	// 	},
	// }
	updateJobInfo := bson.M{}
	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}

	_id, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}

	results := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))
	var jobListing model.JobListing
	if err := results.Decode(&jobListing); err != nil {
		log.Fatal(err)
	}
	return &jobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, err := primitive.ObjectIDFromHex(jobId)
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"_id": _id}
	_, err = jobCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return &model.DeleteJobResponse{DeleteJobID: jobId}
}
