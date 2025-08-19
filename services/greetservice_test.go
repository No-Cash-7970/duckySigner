package services_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "duckysigner/services"
)

var _ = Describe("Greet Service", func() {
	Describe("GreetService.Greet()", func() {
		It("says hello with given name", func() {
			greetService := GreetService{}
			Expect(greetService.Greet("world")).To(Equal("Hello world!"))
		})
	})
})
