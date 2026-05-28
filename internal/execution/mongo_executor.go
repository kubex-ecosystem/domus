package execution

import (
	"context"

	"github.com/kubex-ecosystem/logz"
	"go.mongodb.org/mongo-driver/bson"
	mdb "go.mongodb.org/mongo-driver/v2/mongo"
	mdbOpts "go.mongodb.org/mongo-driver/v2/mongo/options"
	mdbReadPref "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoExecutor expõe as operações básicas do mongo.
type MongoExecutor interface {
	// Exec executa um aggregate.
	Exec(ctx context.Context, db *mdb.Database, col *mdb.Collection, pipeline []bson.M, opts mdbOpts.Lister[mdbOpts.AggregateOptions]) (*mdb.Cursor, error)
	// Query executa uma query.
	Query(ctx context.Context, db *mdb.Database, col *mdb.Collection, filter bson.M, opts mdbOpts.Lister[mdbOpts.FindOptions]) (*mdb.Cursor, error)
	// QueryRow executa uma query e retorna um único resultado.
	QueryRow(ctx context.Context, db *mdb.Database, col *mdb.Collection, filter bson.M, opts mdbOpts.Lister[mdbOpts.FindOneOptions]) (*mdb.SingleResult, error)

	// BeginTx inicia uma transação.
	BeginTx(ctx context.Context, opts mdbOpts.Lister[mdbOpts.SessionOptions]) (*mdb.Session, error)

	// Pool expõe o *mdb.Client em último caso (ex: uso avançado).
	Pool(ctx context.Context, rp *mdbReadPref.ReadPref) (*mdb.Client, error)
}

// mongoExecutor é a implementação padrão baseada em *mdb.Client.
type mongoExecutor struct{ client *mdb.Client }

// NewMongoExecutor cria um MongoExecutor a partir de um *mdb.Client.
func NewMongoExecutor(client *mdb.Client) MongoExecutor {
	return &mongoExecutor{client: client}
}

// Exec executa um aggregate.
func (e *mongoExecutor) Exec(ctx context.Context, db *mdb.Database, col *mdb.Collection, pipeline []bson.M, opts mdbOpts.Lister[mdbOpts.AggregateOptions]) (*mdb.Cursor, error) {
	cs, err := col.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, logz.Errorf("exec failed: %v", err)
	}

	if cs == nil {
		return nil, logz.Errorf("no change stream returned for exec: %v", pipeline)
	}
	return cs, nil
}

// Query executa uma query.
func (e *mongoExecutor) Query(ctx context.Context, db *mdb.Database, col *mdb.Collection, filter bson.M, opts mdbOpts.Lister[mdbOpts.FindOptions]) (*mdb.Cursor, error) {
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, logz.Errorf("query failed: %v", err)
	}
	if cursor == nil {
		return nil, logz.Errorf("no cursor returned for query: %v", filter)
	}
	return cursor, nil
}

// QueryRow executa uma query e retorna um único resultado.
func (e *mongoExecutor) QueryRow(ctx context.Context, db *mdb.Database, col *mdb.Collection, filter bson.M, opts mdbOpts.Lister[mdbOpts.FindOneOptions]) (*mdb.SingleResult, error) {
	r := col.FindOne(ctx, filter, opts)
	if r == nil {
		return nil, logz.Errorf("no row returned for query: %v", filter)
	}
	return r, nil
}

// BeginTx inicia uma transação.
func (e *mongoExecutor) BeginTx(ctx context.Context, opts mdbOpts.Lister[mdbOpts.SessionOptions]) (*mdb.Session, error) {
	return e.client.StartSession(opts)
}

// Pool retorna o *mdb.Client.
func (e *mongoExecutor) Pool(ctx context.Context, rp *mdbReadPref.ReadPref) (*mdb.Client, error) {
	if e.client == nil {
		return nil, logz.Errorf("no mongo client")
	}

	if err := e.client.Ping(ctx, rp); err != nil {
		return nil, logz.Errorf("error pinging mongo client: %v", err)
	}

	return e.client, nil
}
