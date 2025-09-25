package handlers_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GET /", Ordered, func() {
	BeforeAll(func() {
		setUpDcService(rootGetPort, "")
	})

	It("works", func() {
		By("Making request to server")
		resp, err := http.Get("http://localhost:" + rootGetPort)
		Expect(err).NotTo(HaveOccurred())

		By("Processing response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())

		By("Checking response from server")
		Expect(string(body)).To(Equal(`"OK"` + "\n"))
	})

	PIt("works with an authenticated request", func() {
		By("By establishing a session")
		// TODO: Initialize a session and then confirm the session (make 2 http requests)

		By("Making request to server with Authorization header")
		// TODO: Set auth header
		resp, err := http.Get("http://localhost:" + rootGetPort)
		Expect(err).NotTo(HaveOccurred())

		By("Processing response from server")
		body, err := getResponseBody(resp)
		Expect(err).NotTo(HaveOccurred())

		By("Checking response from server")
		Expect(string(body)).To(Equal(`"OK"` + "\n"))
	})
})
