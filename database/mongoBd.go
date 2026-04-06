package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"PReQual/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client
var ctx context.Context

func InitMongoDB(uri string) *mongo.Client {
	ctx = context.Background()
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(fmt.Sprintf("MongoDB connection failed: %v", err))
	}
	return mongoClient
}

func InsertPR(org string, repo string, pr model.PullRequest, headMetrics map[string]interface{}, baseMetrics map[string]interface{}, stats model.AnalysisStat) {
	if mongoClient == nil {
		panic("MongoDB client is not initialized. Call InitMongoDB first.")
	}

	collection := mongoClient.Database("PReQual").Collection("pull_requests")

	// Construire l'ID unique
	prID := fmt.Sprintf("%s_%s_pr%d", org, repo, pr.Number)

	comments := formatComments(pr)
	reviews := formatReviews(pr)

	// Document complet
	prDoc := bson.M{
		"_id":  prID,
		"org":  org,
		"repo": repo,
		"meta": bson.M{
			"id":     pr.Id,
			"number": pr.Number,
			"author": bson.M{
				"login":  pr.Author.Login,
				"is_bot": pr.Author.IsBot,
			},
			"title":      pr.Title,
			"body":       pr.Body,
			"state":      pr.State,
			"created_at": pr.CreatedAt,
			"closed_at":  pr.ClosedAt,
			"merged_at":  pr.MergedAt,
		},
		"head":     headMetrics,
		"base":     baseMetrics,
		"comments": comments,
		"reviews":  reviews,
		"stats": bson.M{
			"total_time": stats.TotalTime,
			"base_size":  stats.BaseSize,
			"head_size":  stats.HeadSize,
		},
		"analysed_at": time.Now(),
	}

	// Upsert (insert ou update si déjà existant)
	filter := bson.M{"_id": prID}
	update := bson.M{"$set": prDoc}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("MongoDB error inserting PR %d: %v", pr.Number, err)
	}
}

func formatReviews(pr model.PullRequest) []bson.M {
	reviews := []bson.M{}

	for _, r := range pr.Reviews {
		reviews = append(reviews, bson.M{
			"author": bson.M{
				"login":  r.Author.Login,
				"is_bot": r.Author.IsBot,
			},
			"state":        r.State,
			"body":         r.Body,
			"submitted_at": r.SubmittedAt,
		})
	}

	return reviews
}

func formatComments(pr model.PullRequest) []bson.M {
	comments := []bson.M{}

	for _, c := range pr.Comments {
		comments = append(comments, bson.M{
			"author": bson.M{
				"login":  c.Author.Login,
				"is_bot": c.Author.IsBot,
			},
			"body":       c.Body,
			"created_at": c.CreatedAt,
		})
	}

	return comments
}
