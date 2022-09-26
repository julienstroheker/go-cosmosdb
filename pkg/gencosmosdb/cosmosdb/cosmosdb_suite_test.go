package cosmosdb_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCosmosdb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cosmosdb Suite")
}
