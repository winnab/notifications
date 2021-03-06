package unsubscribers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/notifications/testing/mocks"
	"github.com/cloudfoundry-incubator/notifications/v2/collections"
	"github.com/cloudfoundry-incubator/notifications/v2/web/unsubscribers"
	"github.com/ryanmoran/stack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteHandler", func() {
	var (
		handler                 unsubscribers.DeleteHandler
		writer                  *httptest.ResponseRecorder
		request                 *http.Request
		context                 stack.Context
		unsubscribersCollection *mocks.UnsubscribersCollection
		database                *mocks.Database
		connection              *mocks.Connection
	)

	BeforeEach(func() {
		var err error

		unsubscribersCollection = mocks.NewUnsubscribersCollection()
		handler = unsubscribers.NewDeleteHandler(unsubscribersCollection)

		connection = mocks.NewConnection()
		database = mocks.NewDatabase()
		database.ConnectionCall.Returns.Connection = connection

		context = stack.NewContext()
		context.Set("database", database)

		writer = httptest.NewRecorder()

		request, err = http.NewRequest("DELETE", "/campaign_types/some-campaign-type-id/unsubscribers/some-user-guid", nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("removes the unsubscribe for the user", func() {
		handler.ServeHTTP(writer, request, context)
		Expect(writer.Code).To(Equal(http.StatusNoContent))
		Expect(writer.Body.String()).To(BeEmpty())

		Expect(unsubscribersCollection.DeleteCall.Receives.Unsubscriber).To(Equal(collections.Unsubscriber{
			CampaignTypeID: "some-campaign-type-id",
			UserGUID:       "some-user-guid",
		}))
		Expect(unsubscribersCollection.DeleteCall.Receives.Connection).To(Equal(connection))
	})

	Context("failure cases", func() {
		Context("when the Delete call returns a NotFoundError", func() {
			It("returns a 404 with the error message in JSON", func() {
				unsubscribersCollection.DeleteCall.Returns.Error = collections.NotFoundError{errors.New("some-error")}
				request, err := http.NewRequest("DELETE", "/campaign_types/some-campaign-type-id/unsubscribers/nonexistent-user", nil)
				Expect(err).NotTo(HaveOccurred())

				handler.ServeHTTP(writer, request, context)

				Expect(writer.Code).To(Equal(http.StatusNotFound))
				Expect(writer.Body.String()).To(MatchJSON(`{"errors": ["some-error"]}`))
			})
		})

		Context("when the Delete call returns a PermissionsError", func() {
			It("returns a 403 status code and reports the error in JSON", func() {
				unsubscribersCollection.DeleteCall.Returns.Error = collections.PermissionsError{errors.New("some-error")}
				request, err := http.NewRequest("DELETE", "/campaign_types/some-campaign-type-id/unsubscribers/some-user-guid", nil)
				Expect(err).NotTo(HaveOccurred())

				handler.ServeHTTP(writer, request, context)

				Expect(writer.Code).To(Equal(http.StatusForbidden))
				Expect(writer.Body.String()).To(MatchJSON(`{"errors": ["some-error"]}`))
			})
		})

		Context("when the Delete call returns any other error", func() {
			It("returns a 500 status", func() {
				unsubscribersCollection.DeleteCall.Returns.Error = errors.New("some-bad-error")
				request, err := http.NewRequest("DELETE", "/campaign_types/some-campaign-type-id/unsubscribers/nonexistent-user", nil)
				Expect(err).NotTo(HaveOccurred())

				handler.ServeHTTP(writer, request, context)

				Expect(writer.Code).To(Equal(http.StatusInternalServerError))
				Expect(writer.Body.String()).To(MatchJSON(`{"errors": ["some-bad-error"]}`))
			})
		})
	})
})
