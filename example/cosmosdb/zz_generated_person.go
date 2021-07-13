// Code generated by github.com/leilifremont/go-cosmosdb, DO NOT EDIT.

package cosmosdb

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	pkg "github.com/leilifremont/go-cosmosdb/example/types"
)

type personClient struct {
	*databaseClient
	path string
}

// PersonClient is a person client
type PersonClient interface {
	Create(context.Context, string, *pkg.Person, *Options) (*pkg.Person, error)
	List(*Options) PersonIterator
	ListAll(context.Context, *Options) (*pkg.People, error)
	Get(context.Context, string, string, *Options) (*pkg.Person, error)
	Replace(context.Context, string, *pkg.Person, *Options) (*pkg.Person, error)
	Delete(context.Context, string, *pkg.Person, *Options) error
	Query(string, *Query, *Options) PersonRawIterator
	QueryAll(context.Context, string, *Query, *Options) (*pkg.People, error)
	ChangeFeed(*Options) PersonIterator
	ExecuteStoredProcedure(ctx context.Context, sprocsid string, partitionKey string, parameters []string) (db *StoredProcedureResponse, err error)
}

type personChangeFeedIterator struct {
	*personClient
	continuation string
	options      *Options
}

type personListIterator struct {
	*personClient
	continuation string
	done         bool
	options      *Options
}

type personQueryIterator struct {
	*personClient
	partitionkey string
	query        *Query
	continuation string
	done         bool
	options      *Options
}

// PersonIterator is a person iterator
type PersonIterator interface {
	Next(context.Context, int) (*pkg.People, error)
	Continuation() string
}

// PersonRawIterator is a person raw iterator
type PersonRawIterator interface {
	PersonIterator
	NextRaw(context.Context, int, interface{}) error
}

// NewPersonClient returns a new person client
func NewPersonClient(collc CollectionClient, collid string) PersonClient {
	return &personClient{
		databaseClient: collc.(*collectionClient).databaseClient,
		path:           collc.(*collectionClient).path + "/colls/" + collid,
	}
}

func (c *personClient) all(ctx context.Context, i PersonIterator) (*pkg.People, error) {
	allpeople := &pkg.People{}

	for {
		people, err := i.Next(ctx, -1)
		if err != nil {
			return nil, err
		}
		if people == nil {
			break
		}

		allpeople.Count += people.Count
		allpeople.ResourceID = people.ResourceID
		allpeople.People = append(allpeople.People, people.People...)
	}

	return allpeople, nil
}

func (c *personClient) Create(ctx context.Context, partitionkey string, newperson *pkg.Person, options *Options) (person *pkg.Person, err error) {
	headers := http.Header{}
	headers.Set("X-Ms-Documentdb-Partitionkey", `["`+partitionkey+`"]`)

	if options == nil {
		options = &Options{}
	}
	options.NoETag = true

	err = c.setOptions(options, newperson, headers)
	if err != nil {
		return
	}

	err = c.do(ctx, http.MethodPost, c.path+"/docs", "docs", c.path, http.StatusCreated, &newperson, &person, headers)
	return
}

// ExecuteStoredProcedure executes a stored procedure in the database
func (c *personClient) ExecuteStoredProcedure(ctx context.Context, sprocsid string, partitionKey string, parameters []string) (db *StoredProcedureResponse, err error) {
	headers := http.Header{}
	headers.Set("X-Ms-documentdb-partitionkey", partitionKey)

	// TODO
	// Double check the request path, request parameters and response are correct
	err = c.do(ctx, http.MethodPost, c.path+"/sprocs/"+sprocsid, "sprocs", c.path+"/sprocs/"+sprocsid, http.StatusOK, &parameters, &db, headers)
	return
}

func (c *personClient) List(options *Options) PersonIterator {
	continuation := ""
	if options != nil {
		continuation = options.Continuation
	}

	return &personListIterator{personClient: c, options: options, continuation: continuation}
}

func (c *personClient) ListAll(ctx context.Context, options *Options) (*pkg.People, error) {
	return c.all(ctx, c.List(options))
}

func (c *personClient) Get(ctx context.Context, partitionkey, personid string, options *Options) (person *pkg.Person, err error) {
	headers := http.Header{}
	headers.Set("X-Ms-Documentdb-Partitionkey", `["`+partitionkey+`"]`)

	err = c.setOptions(options, nil, headers)
	if err != nil {
		return
	}

	err = c.do(ctx, http.MethodGet, c.path+"/docs/"+personid, "docs", c.path+"/docs/"+personid, http.StatusOK, nil, &person, headers)
	return
}

func (c *personClient) Replace(ctx context.Context, partitionkey string, newperson *pkg.Person, options *Options) (person *pkg.Person, err error) {
	headers := http.Header{}
	headers.Set("X-Ms-Documentdb-Partitionkey", `["`+partitionkey+`"]`)

	err = c.setOptions(options, newperson, headers)
	if err != nil {
		return
	}

	err = c.do(ctx, http.MethodPut, c.path+"/docs/"+newperson.ID, "docs", c.path+"/docs/"+newperson.ID, http.StatusOK, &newperson, &person, headers)
	return
}

func (c *personClient) Delete(ctx context.Context, partitionkey string, person *pkg.Person, options *Options) (err error) {
	headers := http.Header{}
	headers.Set("X-Ms-Documentdb-Partitionkey", `["`+partitionkey+`"]`)

	err = c.setOptions(options, person, headers)
	if err != nil {
		return
	}

	err = c.do(ctx, http.MethodDelete, c.path+"/docs/"+person.ID, "docs", c.path+"/docs/"+person.ID, http.StatusNoContent, nil, nil, headers)
	return
}

func (c *personClient) Query(partitionkey string, query *Query, options *Options) PersonRawIterator {
	continuation := ""
	if options != nil {
		continuation = options.Continuation
	}

	return &personQueryIterator{personClient: c, partitionkey: partitionkey, query: query, options: options, continuation: continuation}
}

func (c *personClient) QueryAll(ctx context.Context, partitionkey string, query *Query, options *Options) (*pkg.People, error) {
	return c.all(ctx, c.Query(partitionkey, query, options))
}

func (c *personClient) ChangeFeed(options *Options) PersonIterator {
	continuation := ""
	if options != nil {
		continuation = options.Continuation
	}

	return &personChangeFeedIterator{personClient: c, options: options, continuation: continuation}
}

func (c *personClient) setOptions(options *Options, person *pkg.Person, headers http.Header) error {
	if options == nil {
		return nil
	}

	if person != nil && !options.NoETag {
		if person.ETag == "" {
			return ErrETagRequired
		}
		headers.Set("If-Match", person.ETag)
	}
	if len(options.PreTriggers) > 0 {
		headers.Set("X-Ms-Documentdb-Pre-Trigger-Include", strings.Join(options.PreTriggers, ","))
	}
	if len(options.PostTriggers) > 0 {
		headers.Set("X-Ms-Documentdb-Post-Trigger-Include", strings.Join(options.PostTriggers, ","))
	}
	if len(options.PartitionKeyRangeID) > 0 {
		headers.Set("X-Ms-Documentdb-PartitionKeyRangeID", options.PartitionKeyRangeID)
	}

	return nil
}

func (i *personChangeFeedIterator) Next(ctx context.Context, maxItemCount int) (people *pkg.People, err error) {
	headers := http.Header{}
	headers.Set("A-IM", "Incremental feed")

	headers.Set("X-Ms-Max-Item-Count", strconv.Itoa(maxItemCount))
	if i.continuation != "" {
		headers.Set("If-None-Match", i.continuation)
	}

	err = i.setOptions(i.options, nil, headers)
	if err != nil {
		return
	}

	err = i.do(ctx, http.MethodGet, i.path+"/docs", "docs", i.path, http.StatusOK, nil, &people, headers)
	if IsErrorStatusCode(err, http.StatusNotModified) {
		err = nil
	}
	if err != nil {
		return
	}

	i.continuation = headers.Get("Etag")

	return
}

func (i *personChangeFeedIterator) Continuation() string {
	return i.continuation
}

func (i *personListIterator) Next(ctx context.Context, maxItemCount int) (people *pkg.People, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	headers.Set("X-Ms-Max-Item-Count", strconv.Itoa(maxItemCount))
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.setOptions(i.options, nil, headers)
	if err != nil {
		return
	}

	err = i.do(ctx, http.MethodGet, i.path+"/docs", "docs", i.path, http.StatusOK, nil, &people, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}

func (i *personListIterator) Continuation() string {
	return i.continuation
}

func (i *personQueryIterator) Next(ctx context.Context, maxItemCount int) (people *pkg.People, err error) {
	err = i.NextRaw(ctx, maxItemCount, &people)
	return
}

func (i *personQueryIterator) NextRaw(ctx context.Context, maxItemCount int, raw interface{}) (err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	headers.Set("X-Ms-Max-Item-Count", strconv.Itoa(maxItemCount))
	headers.Set("X-Ms-Documentdb-Isquery", "True")
	headers.Set("Content-Type", "application/query+json")
	if i.partitionkey != "" {
		headers.Set("X-Ms-Documentdb-Partitionkey", `["`+i.partitionkey+`"]`)
	} else {
		headers.Set("X-Ms-Documentdb-Query-Enablecrosspartition", "True")
	}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.setOptions(i.options, nil, headers)
	if err != nil {
		return
	}

	err = i.do(ctx, http.MethodPost, i.path+"/docs", "docs", i.path, http.StatusOK, &i.query, &raw, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}

func (i *personQueryIterator) Continuation() string {
	return i.continuation
}
