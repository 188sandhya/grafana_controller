// +build unitTests

package grafana_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
)

var _ = Describe("Client", func() {
	const testURL = "http://grafana:8000"

	Describe("New()", func() {
		Context("When url and auth header are filled", func() {
			It("Returns client", func() {
				client, err := New(testURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})

		Context("When url can't be parse", func() {
			It("Returns error", func() {
				notParsingURL := `ś://nourl`
				client, err := New(notParsingURL)
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("parse \"%s\": first path segment in URL cannot contain colon", notParsingURL))
			})
		})
	})
})
