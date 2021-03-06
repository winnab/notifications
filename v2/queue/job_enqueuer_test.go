package queue_test

import (
	"bytes"
	"errors"
	"log"
	"time"

	"gopkg.in/gorp.v1"

	"github.com/cloudfoundry-incubator/notifications/cf"
	"github.com/cloudfoundry-incubator/notifications/testing/mocks"
	"github.com/cloudfoundry-incubator/notifications/v2/models"
	"github.com/cloudfoundry-incubator/notifications/v2/queue"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobEnqueuer", func() {
	var (
		enqueuer          queue.JobEnqueuer
		logger            *log.Logger
		buffer            *bytes.Buffer
		gobbleQueue       *mocks.Queue
		gobbleInitializer *mocks.GobbleInitializer
		conn              *mocks.Connection
		transaction       *mocks.Transaction
		messagesRepo      *mocks.MessagesRepository
		space             cf.CloudControllerSpace
		org               cf.CloudControllerOrganization
		reqReceived       time.Time
	)

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		logger = log.New(buffer, "", 0)
		gobbleQueue = mocks.NewQueue()

		transaction = mocks.NewTransaction()
		conn = mocks.NewConnection()

		conn.TransactionCall.Returns.Transaction = transaction
		transaction.Connection = conn
		transaction.GetDbMapCall.Returns.DbMap = &gorp.DbMap{}

		gobbleInitializer = mocks.NewGobbleInitializer()

		messagesRepo = mocks.NewMessagesRepository()
		messagesRepo.InsertCalls = messagesRepo.InsertCalls.WithMessages([]models.Message{
			{
				ID:     "first-random-guid",
				Status: queue.StatusQueued,
			},
			{
				ID:     "second-random-guid",
				Status: queue.StatusQueued,
			},
			{
				ID:     "third-random-guid",
				Status: queue.StatusQueued,
			},
			{
				ID:     "fourth-random-guid",
				Status: queue.StatusQueued,
			},
		})

		enqueuer = queue.NewJobEnqueuer(gobbleQueue, messagesRepo, gobbleInitializer)
		space = cf.CloudControllerSpace{Name: "the-space"}
		org = cf.CloudControllerOrganization{Name: "the-org"}
		reqReceived, _ = time.Parse(time.RFC3339Nano, "2015-06-08T14:40:12.207187819-07:00")
		reqReceived = reqReceived.UTC()
	})

	Describe("Enqueue", func() {
		It("enqueues jobs with the deliveries", func() {
			users := []queue.User{
				{GUID: "user-1", Endorsement: "endores 1"},
				{GUID: "user-2", Endorsement: "endores 2"},
				{GUID: "user-3", Endorsement: "endores 3"},
				{GUID: "user-4", Endorsement: "endores 4"},
			}
			enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

			var deliveries []queue.Delivery
			for _, job := range gobbleQueue.EnqueueCall.Receives.Jobs {
				var delivery queue.Delivery
				err := job.Unmarshal(&delivery)
				if err != nil {
					panic(err)
				}
				deliveries = append(deliveries, delivery)
			}

			Expect(deliveries).To(HaveLen(4))
			Expect(deliveries).To(ConsistOf([]queue.Delivery{
				{
					JobType:         "v2",
					Options:         queue.Options{Endorsement: "endores 1"},
					UserGUID:        "user-1",
					Space:           space,
					Organization:    org,
					ClientID:        "the-client",
					MessageID:       "first-random-guid",
					UAAHost:         "my-uaa-host",
					Scope:           "my.scope",
					VCAPRequestID:   "some-request-id",
					RequestReceived: reqReceived,
					CampaignID:      "some-campaign",
				},
				{
					JobType:         "v2",
					Options:         queue.Options{Endorsement: "endores 2"},
					UserGUID:        "user-2",
					Space:           space,
					Organization:    org,
					ClientID:        "the-client",
					MessageID:       "second-random-guid",
					UAAHost:         "my-uaa-host",
					Scope:           "my.scope",
					VCAPRequestID:   "some-request-id",
					RequestReceived: reqReceived,
					CampaignID:      "some-campaign",
				},
				{
					JobType:         "v2",
					Options:         queue.Options{Endorsement: "endores 3"},
					UserGUID:        "user-3",
					Space:           space,
					Organization:    org,
					ClientID:        "the-client",
					MessageID:       "third-random-guid",
					UAAHost:         "my-uaa-host",
					Scope:           "my.scope",
					VCAPRequestID:   "some-request-id",
					RequestReceived: reqReceived,
					CampaignID:      "some-campaign",
				},
				{
					JobType:         "v2",
					Options:         queue.Options{Endorsement: "endores 4"},
					UserGUID:        "user-4",
					Space:           space,
					Organization:    org,
					ClientID:        "the-client",
					MessageID:       "fourth-random-guid",
					UAAHost:         "my-uaa-host",
					Scope:           "my.scope",
					VCAPRequestID:   "some-request-id",
					RequestReceived: reqReceived,
					CampaignID:      "some-campaign",
				},
			}))
		})

		It("Inserts a StatusQueued for each of the jobs", func() {
			users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
			enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

			var messages []models.Message
			for _, call := range messagesRepo.InsertCalls {
				messages = append(messages, call.Receives.Message)
			}

			Expect(messages).To(HaveLen(4))
			Expect(messages).To(ConsistOf([]models.Message{
				{
					Status:     queue.StatusQueued,
					CampaignID: "some-campaign",
				},
				{
					Status:     queue.StatusQueued,
					CampaignID: "some-campaign",
				},
				{
					Status:     queue.StatusQueued,
					CampaignID: "some-campaign",
				},
				{
					Status:     queue.StatusQueued,
					CampaignID: "some-campaign",
				},
			}))
		})

		Context("using a transaction", func() {
			It("initializes the DbMap", func() {
				users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

				isSamePtr := (gobbleInitializer.InitializeDBMapCall.Receives.DbMap == transaction.GetDbMapCall.Returns.DbMap)
				Expect(isSamePtr).To(BeTrue())
				Expect(transaction.GetDbMapCall.WasCalled).To(BeTrue())
			})

			It("commits the transaction when everything goes well", func() {
				users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

				Expect(transaction.BeginCall.WasCalled).To(BeTrue())
				Expect(transaction.CommitCall.WasCalled).To(BeTrue())
				Expect(transaction.RollbackCall.WasCalled).To(BeFalse())
			})

			It("rolls back the transaction when there is an error in message repo inserting", func() {
				messagesRepo.InsertCalls[0].Returns.Error = errors.New("BOOM!")
				users := []queue.User{{GUID: "user-1"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

				Expect(transaction.BeginCall.WasCalled).To(BeTrue())
				Expect(transaction.CommitCall.WasCalled).To(BeFalse())
				Expect(transaction.RollbackCall.WasCalled).To(BeTrue())
			})

			It("rolls back the transaction when there is an error in enqueuing", func() {
				gobbleQueue.EnqueueCall.Returns.Error = errors.New("BOOM!")
				users := []queue.User{{GUID: "user-1"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

				Expect(transaction.BeginCall.WasCalled).To(BeTrue())
				Expect(transaction.CommitCall.WasCalled).To(BeFalse())
				Expect(transaction.RollbackCall.WasCalled).To(BeTrue())
			})

			Context("when the commit fails", func() {
				It("does not call Rollback", func() {
					transaction.CommitCall.Returns.Error = errors.New("the commit blew up")
					users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
					enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

					Expect(transaction.BeginCall.WasCalled).To(BeTrue())
					Expect(transaction.CommitCall.WasCalled).To(BeTrue())
					Expect(transaction.RollbackCall.WasCalled).To(BeFalse())
				})
			})

			It("uses the same transaction for the queue as it did for the messages repo", func() {
				users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")

				Expect(messagesRepo.InsertCall.Receives.Connection).To(Equal(transaction))
				Expect(gobbleQueue.EnqueueCall.Receives.Connection).To(Equal(transaction))
			})

			It("does not commit the transaction until the jobs have been queued", func() {
				gobbleQueue.EnqueueCall.Hook = func() {
					Expect(transaction.CommitCall.WasCalled).To(BeFalse())
				}

				users := []queue.User{{GUID: "user-1"}, {GUID: "user-2"}, {GUID: "user-3"}, {GUID: "user-4"}}
				enqueuer.Enqueue(conn, users, queue.Options{}, space, org, "the-client", "my-uaa-host", "my.scope", "some-request-id", reqReceived, "some-campaign")
			})
		})
	})
})
