package models_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/postal"
	"github.com/cloudfoundry-incubator/notifications/testing/helpers"
	"github.com/cloudfoundry-incubator/notifications/v1/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MessagesRepo", func() {
	var repo models.MessagesRepo
	var conn db.ConnectionInterface
	var message models.Message

	BeforeEach(func() {
		repo = models.NewMessagesRepo()
		database := db.NewDatabase(sqlDB, db.Config{})
		helpers.TruncateTables(database)
		conn = database.Connection()
		message = models.Message{
			ID:         "message-id-123",
			CampaignID: "some-campaign-id",
			Status:     postal.StatusDelivered,
		}

	})

	Describe("FindByID", func() {
		It("finds messages created in the database", func() {
			message, err := repo.Create(conn, message)
			Expect(err).NotTo(HaveOccurred())

			messageFound, err := repo.FindByID(conn, message.ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(messageFound).To(Equal(message))
		})

		Context("When the message does not exists", func() {
			It("FindByID returns a models.RecordNotFoundError", func() {
				_, err := repo.FindByID(conn, "missing-id")
				Expect(err).To(MatchError(models.NotFoundError{errors.New("Message with ID \"missing-id\" could not be found")}))
			})
		})
	})

	Describe("Upsert", func() {
		Context("when no record exists yet with the message id", func() {
			It("inserts a new record", func() {
				message.UpdatedAt = time.Now().Add(100 * time.Hour)
				_, err := repo.Upsert(conn, message)
				Expect(err).NotTo(HaveOccurred())

				messageFound, err := repo.FindByID(conn, message.ID)
				Expect(err).ToNot(HaveOccurred())

				Expect(messageFound.ID).To(Equal(message.ID))
				Expect(messageFound.Status).To(Equal(message.Status))
				Expect(messageFound.UpdatedAt).To(BeTemporally("~", time.Now(), time.Minute))
			})
		})
		Context("when a record already exists with the message id", func() {
			It("updates the existing record", func() {
				_, err := repo.Create(conn, message)
				Expect(err).NotTo(HaveOccurred())

				updatedMessage := message
				updatedMessage.Status = postal.StatusFailed

				updatedMessage.UpdatedAt = time.Now().Add(100 * time.Hour)
				updatedMessage, err = repo.Upsert(conn, updatedMessage)
				Expect(err).NotTo(HaveOccurred())

				messageFound, err := repo.FindByID(conn, message.ID)
				Expect(err).ToNot(HaveOccurred())

				Expect(messageFound.UpdatedAt).To(BeTemporally("~", time.Now(), time.Minute))
				Expect(messageFound.ID).To(Equal(updatedMessage.ID))
				Expect(messageFound.Status).To(Equal(updatedMessage.Status))
			})
		})
	})

	Describe("DeleteBefore", func() {
		It("Deletes messages older than the input time", func() {
			_, err := repo.Create(conn, message)
			Expect(err).NotTo(HaveOccurred())

			itemsDeleted, err := repo.DeleteBefore(conn, time.Now().Add(1*time.Hour))
			Expect(err).ToNot(HaveOccurred())
			Expect(itemsDeleted).To(Equal(1))

			_, err = repo.FindByID(conn, message.ID)
			Expect(err).To(MatchError(models.NotFoundError{fmt.Errorf("Message with ID %q could not be found", message.ID)}))

		})

		It("Does not delete messages younger than the input time", func() {
			_, err := repo.Create(conn, message)
			Expect(err).NotTo(HaveOccurred())

			itemsDeleted, err := repo.DeleteBefore(conn, time.Now().Add(-1*time.Hour))
			Expect(err).ToNot(HaveOccurred())
			Expect(itemsDeleted).To(Equal(0))

			_, err = repo.FindByID(conn, message.ID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
