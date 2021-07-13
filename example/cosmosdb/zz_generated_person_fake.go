// Code generated by github.com/leilifremont/go-cosmosdb, DO NOT EDIT.

package cosmosdb

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/ugorji/go/codec"

	pkg "github.com/leilifremont/go-cosmosdb/example/types"
)

type fakePersonTriggerHandler func(context.Context, *pkg.Person) error
type fakePersonQueryHandler func(PersonClient, *Query, *Options) PersonRawIterator

var _ PersonClient = &FakePersonClient{}

// NewFakePersonClient returns a FakePersonClient
func NewFakePersonClient(h *codec.JsonHandle) *FakePersonClient {
	return &FakePersonClient{
		jsonHandle:      h,
		people:          make(map[string]*pkg.Person),
		triggerHandlers: make(map[string]fakePersonTriggerHandler),
		queryHandlers:   make(map[string]fakePersonQueryHandler),
	}
}

// FakePersonClient is a FakePersonClient
type FakePersonClient struct {
	lock            sync.RWMutex
	jsonHandle      *codec.JsonHandle
	people          map[string]*pkg.Person
	triggerHandlers map[string]fakePersonTriggerHandler
	queryHandlers   map[string]fakePersonQueryHandler
	sorter          func([]*pkg.Person)
	etag            int

	// returns true if documents conflict
	conflictChecker func(*pkg.Person, *pkg.Person) bool

	// err, if not nil, is an error to return when attempting to communicate
	// with this Client
	err error
}

// SetError sets or unsets an error that will be returned on any
// FakePersonClient method invocation
func (c *FakePersonClient) SetError(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.err = err
}

// SetSorter sets or unsets a sorter function which will be used to sort values
// returned by List() for test stability
func (c *FakePersonClient) SetSorter(sorter func([]*pkg.Person)) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.sorter = sorter
}

// SetConflictChecker sets or unsets a function which can be used to validate
// additional unique keys in a Person
func (c *FakePersonClient) SetConflictChecker(conflictChecker func(*pkg.Person, *pkg.Person) bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.conflictChecker = conflictChecker
}

// SetTriggerHandler sets or unsets a trigger handler
func (c *FakePersonClient) SetTriggerHandler(triggerName string, trigger fakePersonTriggerHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.triggerHandlers[triggerName] = trigger
}

// SetQueryHandler sets or unsets a query handler
func (c *FakePersonClient) SetQueryHandler(queryName string, query fakePersonQueryHandler) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.queryHandlers[queryName] = query
}

func (c *FakePersonClient) deepCopy(person *pkg.Person) (*pkg.Person, error) {
	var b []byte
	err := codec.NewEncoderBytes(&b, c.jsonHandle).Encode(person)
	if err != nil {
		return nil, err
	}

	person = nil
	err = codec.NewDecoderBytes(b, c.jsonHandle).Decode(&person)
	if err != nil {
		return nil, err
	}

	return person, nil
}

func (c *FakePersonClient) apply(ctx context.Context, partitionkey string, person *pkg.Person, options *Options, isCreate bool) (*pkg.Person, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	person, err := c.deepCopy(person) // copy now because pretriggers can mutate person
	if err != nil {
		return nil, err
	}

	if options != nil {
		err := c.processPreTriggers(ctx, person, options)
		if err != nil {
			return nil, err
		}
	}

	existingPerson, exists := c.people[person.ID]
	if isCreate && exists {
		return nil, &Error{
			StatusCode: http.StatusConflict,
			Message:    "Entity with the specified id already exists in the system",
		}
	}
	if !isCreate {
		if !exists {
			return nil, &Error{StatusCode: http.StatusNotFound}
		}

		if person.ETag != existingPerson.ETag {
			return nil, &Error{StatusCode: http.StatusPreconditionFailed}
		}
	}

	if c.conflictChecker != nil {
		for _, personToCheck := range c.people {
			if c.conflictChecker(personToCheck, person) {
				return nil, &Error{
					StatusCode: http.StatusConflict,
					Message:    "Entity with the specified id already exists in the system",
				}
			}
		}
	}

	person.ETag = fmt.Sprint(c.etag)
	c.etag++

	c.people[person.ID] = person

	return c.deepCopy(person)
}

// Create creates a Person in the database
func (c *FakePersonClient) Create(ctx context.Context, partitionkey string, person *pkg.Person, options *Options) (*pkg.Person, error) {
	return c.apply(ctx, partitionkey, person, options, true)
}

// ExecuteStoredProcedure executes a stored procedure in the database
func (c *FakePersonClient) ExecuteStoredProcedure(ctx context.Context, sprocsid string, partitionKey string, parameters []string) (db *StoredProcedureResponse, err error) {
	headers := http.Header{}
	headers.Set("X-Ms-documentdb-partitionkey", partitionKey)

	// TODO
	// Find out what should we do here for fake person？It seems not used in our code
	return
}

// Replace replaces a Person in the database
func (c *FakePersonClient) Replace(ctx context.Context, partitionkey string, person *pkg.Person, options *Options) (*pkg.Person, error) {
	return c.apply(ctx, partitionkey, person, options, false)
}

// List returns a PersonIterator to list all People in the database
func (c *FakePersonClient) List(*Options) PersonIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakePersonErroringRawIterator(c.err)
	}

	people := make([]*pkg.Person, 0, len(c.people))
	for _, person := range c.people {
		person, err := c.deepCopy(person)
		if err != nil {
			return NewFakePersonErroringRawIterator(err)
		}
		people = append(people, person)
	}

	if c.sorter != nil {
		c.sorter(people)
	}

	return NewFakePersonIterator(people, 0)
}

// ListAll lists all People in the database
func (c *FakePersonClient) ListAll(ctx context.Context, options *Options) (*pkg.People, error) {
	iter := c.List(options)
	return iter.Next(ctx, -1)
}

// Get gets a Person from the database
func (c *FakePersonClient) Get(ctx context.Context, partitionkey string, id string, options *Options) (*pkg.Person, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return nil, c.err
	}

	person, exists := c.people[id]
	if !exists {
		return nil, &Error{StatusCode: http.StatusNotFound}
	}

	return c.deepCopy(person)
}

// Delete deletes a Person from the database
func (c *FakePersonClient) Delete(ctx context.Context, partitionKey string, person *pkg.Person, options *Options) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.err != nil {
		return c.err
	}

	_, exists := c.people[person.ID]
	if !exists {
		return &Error{StatusCode: http.StatusNotFound}
	}

	delete(c.people, person.ID)
	return nil
}

// ChangeFeed is unimplemented
func (c *FakePersonClient) ChangeFeed(*Options) PersonIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakePersonErroringRawIterator(c.err)
	}

	return NewFakePersonErroringRawIterator(ErrNotImplemented)
}

func (c *FakePersonClient) processPreTriggers(ctx context.Context, person *pkg.Person, options *Options) error {
	for _, triggerName := range options.PreTriggers {
		if triggerHandler := c.triggerHandlers[triggerName]; triggerHandler != nil {
			c.lock.Unlock()
			err := triggerHandler(ctx, person)
			c.lock.Lock()
			if err != nil {
				return err
			}
		} else {
			return ErrNotImplemented
		}
	}

	return nil
}

// Query calls a query handler to implement database querying
func (c *FakePersonClient) Query(name string, query *Query, options *Options) PersonRawIterator {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.err != nil {
		return NewFakePersonErroringRawIterator(c.err)
	}

	if queryHandler := c.queryHandlers[query.Query]; queryHandler != nil {
		c.lock.RUnlock()
		i := queryHandler(c, query, options)
		c.lock.RLock()
		return i
	}

	return NewFakePersonErroringRawIterator(ErrNotImplemented)
}

// QueryAll calls a query handler to implement database querying
func (c *FakePersonClient) QueryAll(ctx context.Context, partitionkey string, query *Query, options *Options) (*pkg.People, error) {
	iter := c.Query("", query, options)
	return iter.Next(ctx, -1)
}

func NewFakePersonIterator(people []*pkg.Person, continuation int) PersonRawIterator {
	return &fakePersonIterator{people: people, continuation: continuation}
}

type fakePersonIterator struct {
	people       []*pkg.Person
	continuation int
	done         bool
}

func (i *fakePersonIterator) NextRaw(ctx context.Context, maxItemCount int, out interface{}) error {
	return ErrNotImplemented
}

func (i *fakePersonIterator) Next(ctx context.Context, maxItemCount int) (*pkg.People, error) {
	if i.done {
		return nil, nil
	}

	var people []*pkg.Person
	if maxItemCount == -1 {
		people = i.people[i.continuation:]
		i.continuation = len(i.people)
		i.done = true
	} else {
		max := i.continuation + maxItemCount
		if max > len(i.people) {
			max = len(i.people)
		}
		people = i.people[i.continuation:max]
		i.continuation += max
		i.done = i.Continuation() == ""
	}

	return &pkg.People{
		People: people,
		Count:  len(people),
	}, nil
}

func (i *fakePersonIterator) Continuation() string {
	if i.continuation >= len(i.people) {
		return ""
	}
	return fmt.Sprintf("%d", i.continuation)
}

// NewFakePersonErroringRawIterator returns a PersonRawIterator which
// whose methods return the given error
func NewFakePersonErroringRawIterator(err error) PersonRawIterator {
	return &fakePersonErroringRawIterator{err: err}
}

type fakePersonErroringRawIterator struct {
	err error
}

func (i *fakePersonErroringRawIterator) Next(ctx context.Context, maxItemCount int) (*pkg.People, error) {
	return nil, i.err
}

func (i *fakePersonErroringRawIterator) NextRaw(context.Context, int, interface{}) error {
	return i.err
}

func (i *fakePersonErroringRawIterator) Continuation() string {
	return ""
}
