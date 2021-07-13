package cosmosdb

import (
	"context"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
)

// Database represents a database
type Database struct {
	ID          string `json:"id,omitempty"`
	ResourceID  string `json:"_rid,omitempty"`
	Timestamp   int    `json:"_ts,omitempty"`
	Self        string `json:"_self,omitempty"`
	ETag        string `json:"_etag,omitempty"`
	Collections string `json:"_colls,omitempty"`
	Users       string `json:"_users,omitempty"`
}

// Databases represents databases
type Databases struct {
	Count      int         `json:"_count,omitempty"`
	ResourceID string      `json:"_rid,omitempty"`
	Databases  []*Database `json:"Databases,omitempty"`
}

// StoredProcedureResponse represents a Stored Procedure Response
type StoredProcedureResponse struct {
	Response string `json:"response,omitempty"`
}

type databaseClient struct {
	mu               sync.RWMutex
	log              *logrus.Entry
	hc               *http.Client
	jsonHandle       *codec.JsonHandle
	databaseHostname string
	authorizer       Authorizer
	maxRetries       int
}

// DatabaseClient is a database client
type DatabaseClient interface {
	SetAuthorizer(Authorizer)
	Create(context.Context, *Database) (*Database, error)
	List() DatabaseIterator
	ListAll(context.Context) (*Databases, error)
	Get(context.Context, string) (*Database, error)
	Delete(context.Context, *Database) error
}

type databaseListIterator struct {
	*databaseClient
	continuation string
	done         bool
}

// DatabaseIterator is a database iterator
type DatabaseIterator interface {
	Next(context.Context) (*Databases, error)
}

// NewDatabaseClient returns a new database client
func NewDatabaseClient(log *logrus.Entry, hc *http.Client, jsonHandle *codec.JsonHandle, databaseHostname string, authorizer Authorizer) DatabaseClient {
	return &databaseClient{
		log:              log,
		hc:               hc,
		jsonHandle:       jsonHandle,
		databaseHostname: databaseHostname,
		authorizer:       authorizer,
		maxRetries:       10,
	}
}

func (c *databaseClient) all(ctx context.Context, i DatabaseIterator) (*Databases, error) {
	alldbs := &Databases{}

	for {
		dbs, err := i.Next(ctx)
		if err != nil {
			return nil, err
		}
		if dbs == nil {
			break
		}

		alldbs.Count += dbs.Count
		alldbs.ResourceID = dbs.ResourceID
		alldbs.Databases = append(alldbs.Databases, dbs.Databases...)
	}

	return alldbs, nil
}

func (c *databaseClient) SetAuthorizer(authorizer Authorizer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.authorizer = authorizer
}

func (c *databaseClient) Create(ctx context.Context, newdb *Database) (db *Database, err error) {
	err = c.do(ctx, http.MethodPost, "dbs", "dbs", "", http.StatusCreated, &newdb, &db, nil)
	return
}

func (c *databaseClient) List() DatabaseIterator {
	return &databaseListIterator{databaseClient: c}
}

func (c *databaseClient) ListAll(ctx context.Context) (*Databases, error) {
	return c.all(ctx, c.List())
}

func (c *databaseClient) Get(ctx context.Context, dbid string) (db *Database, err error) {
	err = c.do(ctx, http.MethodGet, "dbs/"+dbid, "dbs", "dbs/"+dbid, http.StatusOK, nil, &db, nil)
	return
}

func (c *databaseClient) Delete(ctx context.Context, db *Database) error {
	if db.ETag == "" {
		return ErrETagRequired
	}
	headers := http.Header{}
	headers.Set("If-Match", db.ETag)
	return c.do(ctx, http.MethodDelete, "dbs/"+db.ID, "dbs", "dbs/"+db.ID, http.StatusNoContent, nil, nil, headers)
}

func (i *databaseListIterator) Next(ctx context.Context) (dbs *Databases, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.do(ctx, http.MethodGet, "dbs", "dbs", "", http.StatusOK, nil, &dbs, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}
