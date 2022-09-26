package cosmosdb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Cosmos DB client utilities", func() {
	It("Should retry on 'http: request canceled' error message", func() {
		callCount := 0
		err := RetryOnHttpStatusOrError(func() error {
			callCount += 1
			if callCount != 5 {
				return &Error{StatusCode: 500, Message: "http: request canceled"}
			}
			return &Error{StatusCode: 404, Message: "Resource Not Found"}
		}, 504, "http: request canceled")
		Expect(callCount).To(Equal(5))
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("404 : Resource Not Found"))
	})

	It("Should not retry on 'Unauthorized' error message", func() {
		callCount := 0
		err := RetryOnHttpStatusOrError(func() error {
			callCount += 1
			if callCount != 5 {
				return &Error{StatusCode: 401, Message: "Unauthorized"}
			}
			return &Error{StatusCode: 500, Message: "Server Error"}
		}, 504, "http: request canceled")
		Expect(callCount).To(Equal(1))
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(Equal("401 : Unauthorized"))
	})
})
