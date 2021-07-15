package cosmosdb

import (
	"context"
	"net/http"
)

// StoredProcedure represents a stored procedure
type StoredProcedure struct {
	// ID is the user generated unique name for the stored procedure. No two stored procedures can have the same IDs.
	ID         string `json:"id,omitempty"`
	ResourceID string `json:"_rid,omitempty"`
	Timestamp  int    `json:"_ts,omitempty"`
	Self       string `json:"_self,omitempty"`
	ETag       string `json:"_etag,omitempty"`
	// Body of the stored procedure.
	Body string `json:"body,omitempty"`
}

// StoredProcedures represents stored procedures
type StoredProcedures struct {
	Count            int                `json:"_count,omitempty"`
	ResourceID       string             `json:"_rid,omitempty"`
	StoredProcedures []*StoredProcedure `json:"StoredProcedures,omitempty"`
}

type storedProcedureClient struct {
	*databaseClient
	path string
}

// StoredProcedureClient is a stored procedure client
type StoredProcedureClient interface {
	Create(context.Context, *StoredProcedure) (*StoredProcedure, error)
	List() StoredProcedureIterator
	ListAll(context.Context) (*StoredProcedures, error)
	Get(context.Context, string) (*StoredProcedure, error)
	Delete(context.Context, *StoredProcedure) error
	Replace(context.Context, *StoredProcedure) (*StoredProcedure, error)
}

type storedProcedureListIterator struct {
	*storedProcedureClient
	continuation string
	done         bool
}

// StoredProcedureIterator is a stored procedure iterator
type StoredProcedureIterator interface {
	Next(context.Context) (*StoredProcedures, error)
}

// NewStoredProcedureClient returns a new stored procedure client
func NewStoredProcedureClient(collc CollectionClient, collid string) StoredProcedureClient {
	return &storedProcedureClient{
		databaseClient: collc.(*collectionClient).databaseClient,
		path:           collc.(*collectionClient).path + "/colls/" + collid,
	}
}

func (c *storedProcedureClient) all(ctx context.Context, i StoredProcedureIterator) (*StoredProcedures, error) {
	allstoredprocedures := &StoredProcedures{}

	for {
		storedprocedures, err := i.Next(ctx)
		if err != nil {
			return nil, err
		}
		if storedprocedures == nil {
			break
		}

		allstoredprocedures.Count += storedprocedures.Count
		allstoredprocedures.ResourceID = storedprocedures.ResourceID
		allstoredprocedures.StoredProcedures = append(allstoredprocedures.StoredProcedures, storedprocedures.StoredProcedures...)
	}

	return allstoredprocedures, nil
}

func (c *storedProcedureClient) Create(ctx context.Context, newstoredprocedure *StoredProcedure) (storedprocedure *StoredProcedure, err error) {
	err = c.do(ctx, http.MethodPost, c.path+"/sprocs", "sprocs", c.path, http.StatusCreated, &newstoredprocedure, &storedprocedure, nil)
	return
}

func (c *storedProcedureClient) List() StoredProcedureIterator {
	return &storedProcedureListIterator{storedProcedureClient: c}
}

func (c *storedProcedureClient) ListAll(ctx context.Context) (*StoredProcedures, error) {
	return c.all(ctx, c.List())
}

func (c *storedProcedureClient) Get(ctx context.Context, storedprocedureid string) (storedprocedure *StoredProcedure, err error) {
	err = c.do(ctx, http.MethodGet, c.path+"/sprocs/"+storedprocedureid, "sprocs", c.path+"/sprocs/"+storedprocedureid, http.StatusOK, nil, &storedprocedure, nil)
	return
}

func (c *storedProcedureClient) Delete(ctx context.Context, storedprocedure *StoredProcedure) error {
	if storedprocedure.ETag == "" {
		return ErrETagRequired
	}
	headers := http.Header{}
	headers.Set("If-Match", storedprocedure.ETag)
	return c.do(ctx, http.MethodDelete, c.path+"/sprocs/"+storedprocedure.ID, "sprocs", c.path+"/sprocs/"+storedprocedure.ID, http.StatusNoContent, nil, nil, headers)
}

func (c *storedProcedureClient) Replace(ctx context.Context, newstoredprocedure *StoredProcedure) (storedprocedure *StoredProcedure, err error) {
	err = c.do(ctx, http.MethodPost, c.path+"/sprocs/"+newstoredprocedure.ID, "sprocs", c.path+"/sprocs/"+newstoredprocedure.ID, http.StatusCreated, &newstoredprocedure, &storedprocedure, nil)
	return
}

func (i *storedProcedureListIterator) Next(ctx context.Context) (storedprocedures *StoredProcedures, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.do(ctx, http.MethodGet, i.path+"/sprocs", "sprocs", i.path, http.StatusOK, nil, &storedprocedures, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}
