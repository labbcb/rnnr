package server

import (
	"context"
	"time"

	"github.com/labbcb/rnnr/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// TaskCollection is the collection name for tasks
	TaskCollection = "tasks"
	// NodeCollection is the collection name for nodes
	NodeCollection = "nodes"
)

// DB wraps MongoDB client to provides task- and node-related operations.
type DB struct {
	client   *mongo.Client
	database string
}

// MongoConnect creates a MongoDB client.
func MongoConnect(uri, database string) (*DB, error) {
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &DB{client: c, database: database}, nil
}

// SaveTask stores a task.
// It will set Task.Created and Task.Updated to current local time.
func (d *DB) SaveTask(t *models.Task) error {
	now := time.Now()
	t.Created = &now
	_, err := d.client.Database(d.database).Collection(TaskCollection).InsertOne(context.Background(), t)
	return err
}

// GetTask finds a task by its ID.
func (d *DB) GetTask(id string) (*models.Task, error) {
	var t models.Task
	if err := d.client.Database(d.database).Collection(TaskCollection).FindOne(context.Background(), bson.M{"_id": id}).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask saves task changes in database.
// It will set Task.Updated to current local time.
func (d *DB) UpdateTask(t *models.Task) error {
	return d.client.Database(d.database).Collection(TaskCollection).
		FindOneAndReplace(context.Background(), bson.M{"_id": t.ID}, &t, options.FindOneAndReplace()).Err()
}

// ListTasks retrieves tasks that match given worker nodes and states.
// Pagination is done via limit and skip parameters.
// view defines task fields to be returned.
//
// Minimal returns only task ID and state.
//
// Basic returns all fields except Logs.ExecutorLogs.Stdout, Logs.ExecutorLogs.Stderr, Inputs.Content and Logs.SystemLogs.
//
// Full returns all fields.
func (d *DB) ListTasks(limit, skip int64, view models.View, nodes []string, states []models.State) ([]*models.Task, error) {
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	opts.SetSkip(skip)

	var filters bson.A
	if len(nodes) != 0 {
		filters = append(filters, bson.M{"worker.host": bson.M{"$in": nodes}})
	}
	if len(states) != 0 {
		filters = append(filters, bson.M{"state": bson.M{"$in": states}})
	}

	filter := bson.M{}
	if len(filters) > 0 {
		filter = bson.M{"$and": filters}
	}

	var projection bson.M
	switch view {
	case models.Minimal:
		projection = bson.M{
			"_id":   1,
			"state": 1,
		}
	case models.Basic:
		projection = bson.M{
			"Logs.ExecutorLog.Stdout": 0,
			"Logs.ExecutorLog.Stderr": 0,
			"Input.Content":           0,
			"Logs.SystemLogs":         0,
		}
	}
	opts.SetProjection(projection)

	cursor, err := d.client.Database(d.database).Collection(TaskCollection).Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	var tasks []*models.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// ListNodes returns worker nodes (disabled included).
// Set active to return active (enabled) or disable nodes.
func (d *DB) ListNodes(active *bool) ([]*models.Node, error) {
	var filter bson.M
	if active != nil {
		filter = bson.M{"active": active}
	}

	cursor, err := d.client.Database(d.database).Collection(NodeCollection).Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	var ns []*models.Node
	if err := cursor.All(context.Background(), &ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// GetNode retrieves a computing node by its server address.
func (d *DB) GetNode(host string) (*models.Node, error) {
	var n models.Node
	if err := d.client.Database(d.database).Collection(NodeCollection).
		FindOne(context.Background(), bson.M{"_id": host}).Decode(&n); err != nil {
		return nil, err
	}
	return &n, nil
}

// AddNode activates a node. If already registered it updates node fields with same ID.
func (d *DB) AddNode(n *models.Node) error {
	_, err := d.GetNode(n.Host)

	switch err {
	case nil:
		return d.UpdateNode(n)
	case mongo.ErrNoDocuments:
		_, err := d.client.Database(d.database).Collection(NodeCollection).InsertOne(context.Background(), n)
		return err
	default:
		return err
	}
}

// UpdateNode updates node information.
func (d *DB) UpdateNode(n *models.Node) error {
	return d.client.Database(d.database).Collection(NodeCollection).
		FindOneAndReplace(context.Background(), bson.M{"_id": n.Host}, n, options.FindOneAndReplace()).Err()
}
