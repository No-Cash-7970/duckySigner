package services_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "duckysigner/services"
)

var _ = Describe("Greetservice", func() {
	Describe("Greet()", func() {
		It("says hello with given name", func() {
			greetService := GreetService{}
			Expect(greetService.Greet("world")).To(Equal("Hello world!"))
		})
	})
})
