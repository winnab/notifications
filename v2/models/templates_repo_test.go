package models_test

import (
	"github.com/cloudfoundry-incubator/notifications/db"
	"github.com/cloudfoundry-incubator/notifications/testing/mocks"
	"github.com/cloudfoundry-incubator/notifications/testing/helpers"
	"github.com/cloudfoundry-incubator/notifications/v2/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TemplatesRepo", func() {
	var (
		repo models.TemplatesRepository
		conn db.ConnectionInterface
	)

	BeforeEach(func() {
		database := db.NewDatabase(sqlDB, db.Config{})
		helpers.TruncateTables(database)
		repo = models.NewTemplatesRepository(mocks.NewIncrementingGUIDGenerator().Generate)
		conn = database.Connection()
	})

	Describe("Insert", func() {
		It("returns the data", func() {
			createdTemplate, err := repo.Insert(conn, models.Template{
				Name:     "some-template",
				ClientID: "some-client-id",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(createdTemplate.ID).To(Equal("deadbeef-aabb-ccdd-eeff-001122334455"))
		})

		Context("failure cases", func() {
			It("returns an error if it happens", func() {
				_, err := repo.Insert(conn, models.Template{
					Name:     "some-template",
					ClientID: "some-client-id",
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = repo.Insert(conn, models.Template{
					Name:     "some-template",
					ClientID: "some-client-id",
				})
				Expect(err).To(BeAssignableToTypeOf(models.DuplicateRecordError{}))
			})
		})
	})

	Describe("Get", func() {
		It("fetches the template given a template_id", func() {
			createdTemplate, err := repo.Insert(conn, models.Template{
				Name:     "some-template",
				ClientID: "some-client-id",
			})
			Expect(err).NotTo(HaveOccurred())

			template, err := repo.Get(conn, createdTemplate.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(template).To(Equal(createdTemplate))
		})

		Context("failure cases", func() {
			It("returns not found error if it happens", func() {
				_, err := repo.Get(conn, "missing-template-id")
				Expect(err).To(BeAssignableToTypeOf(models.RecordNotFoundError{}))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the template given a template_id", func() {
			template, err := repo.Insert(conn, models.Template{
				Name:     "some-template",
				ClientID: "some-client-id",
			})
			Expect(err).NotTo(HaveOccurred())

			err = repo.Delete(conn, template.ID)
			Expect(err).NotTo(HaveOccurred())

			_, err = repo.Get(conn, template.ID)
			Expect(err).To(BeAssignableToTypeOf(models.RecordNotFoundError{}))
		})

		Context("failure cases", func() {
			It("returns not found error if it happens", func() {
				err := repo.Delete(conn, "missing-template-id")
				Expect(err).To(BeAssignableToTypeOf(models.RecordNotFoundError{}))
			})
		})
	})
})