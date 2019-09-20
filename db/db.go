package db

import (
	"github.com/labbcb/rnnr/node"
	"github.com/labbcb/rnnr/task"
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

// Save stores a task
func (d *DB) Save(t *task.Task) error {
	_, err := d.client.Database(d.database).Collection(TaskCollection).InsertOne(nil, t)
	return err
}

// Get finds a task by its ID
func (d *DB) Get(id string) (*task.Task, error) {
	var t task.Task
	if err := d.client.Database(d.database).Collection(TaskCollection).FindOne(nil, bson.M{"_id": id}).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Update saves task changes in database
func (d *DB) Update(t *task.Task) error {
	return d.client.Database(d.database).Collection(TaskCollection).
		FindOneAndReplace(nil, bson.M{"_id": t.ID}, &t, options.FindOneAndReplace()).Err()
}

// FindByState returns tasks that matches states
func (d *DB) FindByState(states ...task.State) ([]*task.Task, error) {
	var ts []*task.Task
	filter := bson.M{"state": bson.M{"$in": states}}
	cur, err := d.client.Database(d.database).Collection(TaskCollection).Find(nil, filter, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(nil)
	if err := cur.All(nil, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// FindAll returns all tasks stored in database
func (d *DB) FindAll() ([]*task.Task, error) {
	var ts []*task.Task
	cur, err := d.client.Database(d.database).Collection(TaskCollection).Find(nil, bson.D{{}}, options.Find())
	defer cur.Close(nil)
	if err != nil {
		return nil, err
	}
	if err := cur.All(nil, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// All returns all nodes
func (d *DB) All() ([]*node.Node, error) {
	cur, err := d.client.Database(d.database).Collection(NodeCollection).Find(nil, bson.D{{}}, options.Find())
	defer cur.Close(nil)
	if err != nil {
		return nil, err
	}
	var ns []*node.Node
	if err := cur.All(nil, &ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// GetByHost retrieves a computing node by its server address.
func (d *DB) GetByHost(host string) (*node.Node, error) {
	var n node.Node
	if err := d.client.Database(d.database).Collection(NodeCollection).
		FindOne(nil, bson.M{"host": host}).Decode(&n); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetByID retrieves a computing node by its ID.
func (d *DB) GetByID(host string) (*node.Node, error) {
	var n node.Node
	if err := d.client.Database(d.database).Collection(NodeCollection).
		FindOne(nil, bson.M{"_id": host}).Decode(&n); err != nil {
		return nil, err
	}
	return &n, nil
}

// Add activates a node. If already registered it updates node fields with same ID.
func (d *DB) Add(n *node.Node) error {
	existing, err := d.GetByHost(n.Host)

	switch err {
	case nil:
		n.ID = existing.ID
		return d.UpdateNode(n)
	case mongo.ErrNoDocuments:
		_, err := d.client.Database(d.database).Collection(NodeCollection).InsertOne(nil, n)
		return err
	default:
		return err
	}
}

// UpdateUsage update usage of node.
func (d *DB) UpdateUsage(n *node.Node) error {
	return d.client.Database(d.database).Collection(NodeCollection).
		FindOneAndUpdate(nil, bson.M{"_id": n.ID}, bson.M{"$set": bson.M{"usage": n.Usage}}, options.FindOneAndUpdate()).Err()
}

// Active returns active computing nodes.
func (d *DB) Active() ([]*node.Node, error) {
	cur, err := d.client.Database(d.database).Collection(NodeCollection).Find(nil, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	var ns []*node.Node
	if err := cur.All(nil, &ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// UpdateNode updates node information.
func (d *DB) UpdateNode(n *node.Node) error {
	return d.client.Database(d.database).Collection(NodeCollection).
		FindOneAndReplace(nil, bson.M{"_id": n.ID}, n, options.FindOneAndReplace()).Err()
}
