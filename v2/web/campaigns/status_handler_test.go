package campaigns_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cloudfoundry-incubator/notifications/application"
	"github.com/cloudfoundry-incubator/notifications/testing/helpers"
	"github.com/cloudfoundry-incubator/notifications/testing/mocks"
	"github.com/cloudfoundry-incubator/notifications/v2/collections"
	"github.com/cloudfoundry-incubator/notifications/v2/web/campaigns"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
	"github.com/ryanmoran/stack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Campaign status handler", func() {
	var (
		handler                    campaigns.StatusHandler
		context                    stack.Context
		writer                     *httptest.ResponseRecorder
		request                    *http.Request
		database                   *mocks.Database
		conn                       *mocks.Connection
		campaignStatusesCollection *mocks.CampaignStatusesCollection
	)

	BeforeEach(func() {
		tokenHeader := map[string]interface{}{
			"alg": "FAST",
		}
		tokenClaims := map[string]interface{}{
			"client_id": "some-uaa-client-id",
			"exp":       int64(3404281214),
			"scope":     []string{"notifications.write"},
		}
		token, err := jwt.Parse(helpers.BuildToken(tokenHeader, tokenClaims), func(*jwt.Token) (interface{}, error) {
			return []byte(application.UAAPublicKey), nil
		})
		Expect(err).NotTo(HaveOccurred())

		conn = mocks.NewConnection()
		database = mocks.NewDatabase()
		database.ConnectionCall.Returns.Connection = conn

		context = stack.NewContext()
		context.Set("token", token)
		context.Set("database", database)
		context.Set("client_id", "my-client")

		writer = httptest.NewRecorder()

		campaignStatusesCollection = mocks.NewCampaignStatusesCollection()

		handler = campaigns.NewStatusHandler(campaignStatusesCollection)
	})

	It("gets the status of an existing campaign", func() {
		startTime, err := time.Parse(time.RFC3339, "2015-09-01T12:34:56-07:00")
		Expect(err).NotTo(HaveOccurred())

		completedTime, err := time.Parse(time.RFC3339, "2015-09-01T12:34:58-07:00")
		Expect(err).NotTo(HaveOccurred())

		campaignStatusesCollection.GetCall.Returns.CampaignStatus = collections.CampaignStatus{
			CampaignID:     "some-campaign-id",
			Status:         "completed",
			TotalMessages:  8,
			SentMessages:   6,
			RetryMessages:  0,
			FailedMessages: 2,
			StartTime:      startTime,
			CompletedTime: mysql.NullTime{
				Time:  completedTime,
				Valid: true,
			},
		}

		request, err = http.NewRequest("GET", "/senders/some-sender-id/campaigns/some-campaign-id/status", nil)
		Expect(err).NotTo(HaveOccurred())

		handler.ServeHTTP(writer, request, context)

		Expect(writer.Code).To(Equal(http.StatusOK))
		Expect(writer.Body).To(MatchJSON(`{
			"id": "some-campaign-id",
			"status": "completed",
			"total_messages": 8,
			"sent_messages": 6,
			"retry_messages": 0,
			"failed_messages": 2,
			"start_time": "2015-09-01T12:34:56-07:00",
			"completed_time": "2015-09-01T12:34:58-07:00"
		}`))

		Expect(campaignStatusesCollection.GetCall.Receives.Connection).To(Equal(conn))
		Expect(campaignStatusesCollection.GetCall.Receives.CampaignID).To(Equal("some-campaign-id"))
	})
})