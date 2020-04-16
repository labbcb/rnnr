package db

import (
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

// DB wraps MongoDB client to provides task- and node-related operations
type DB struct {
	client   *mongo.Client
	database string
}

// Connect creates a MongoDB client
func Connect(uri, database string) (*DB, error) {
	c, err := mongo.Connect(nil, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &DB{client: c, database: database}, nil
}

// SaveTask stores a task
func (d *DB) SaveTask(t *models.Task) error {
	_, err := d.client.Database(d.database).Collection(TaskCollection).InsertOne(nil, t)
	return err
}

// GetTask finds a task by its ID
func (d *DB) GetTask(id string) (*models.Task, error) {
	var t models.Task
	if err := d.client.Database(d.database).Collection(TaskCollection).FindOne(nil, bson.M{"_id": id}).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTask saves task changes in database
func (d *DB) UpdateTask(t *models.Task) error {
	return d.client.Database(d.database).Collection(TaskCollection).
		FindOneAndReplace(nil, bson.M{"_id": t.ID}, &t, options.FindOneAndReplace()).Err()
}

// FindByState retrieves from database tasks that match given states.
func (d *DB) FindByState(limit, skip int64, states ...models.State) ([]*models.Task, error) {
	var opts *options.FindOptions
	if limit > 0 {
		opts = &options.FindOptions{Limit: &limit, Skip: &skip}
	} else {
		opts = &options.FindOptions{Limit: nil, Skip: &skip}
	}

	var filter bson.M
	if len(states) != 0 {
		filter = bson.M{"state": bson.M{"$in": states}}
	}

	cursor, err := d.client.Database(d.database).Collection(TaskCollection).Find(nil, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(nil); err != nil {
			log.Fatal(err)
		}
	}()

	var tasks []*models.Task
	if err := cursor.All(nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// AllNodes returns all nodes.
func (d *DB) AllNodes() ([]*models.Node, error) {
	cursor, err := d.client.Database(d.database).Collection(NodeCollection).Find(nil, bson.D{{}}, options.Find())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(nil); err != nil {
			log.Fatal(err)
		}
	}()

	var ns []*models.Node
	if err := cursor.All(nil, &ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// GetNode retrieves a computing node by its server address.
func (d *DB) GetNode(host string) (*models.Node, error) {
	var n models.Node
	if err := d.client.Database(d.database).Collection(NodeCollection).
		FindOne(nil, bson.M{"_id": host}).Decode(&n); err != nil {
		return nil, err
	}
	return &n, nil
}

// AddNodes activates a node. If already registered it updates node fields with same ID.
func (d *DB) AddNodes(n *models.Node) error {
	_, err := d.GetNode(n.Host)

	switch err {
	case nil:
		return d.UpdateNode(n)
	case mongo.ErrNoDocuments:
		_, err := d.client.Database(d.database).Collection(NodeCollection).InsertOne(nil, n)
		return err
	default:
		return err
	}
}

// GetActiveNodes returns active computing nodes.
func (d *DB) GetActiveNodes() ([]*models.Node, error) {
	cur, err := d.client.Database(d.database).Collection(NodeCollection).Find(nil, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	var ns []*models.Node
	if err := cur.All(nil, &ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// UpdateNode updates node information.
func (d *DB) UpdateNode(n *models.Node) error {
	return d.client.Database(d.database).Collection(NodeCollection).
		FindOneAndReplace(nil, bson.M{"_id": n.Host}, n, options.FindOneAndReplace()).Err()
}
