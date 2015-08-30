package services_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/notifications/testing/mocks"
	"github.com/cloudfoundry-incubator/notifications/v1/models"
	"github.com/cloudfoundry-incubator/notifications/v1/services"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finder", func() {
	var (
		finder        services.TemplateFinder
		templatesRepo *mocks.TemplatesRepo
		database      *mocks.Database
		conn          *mocks.Connection
	)

	Describe("#FindByID", func() {
		BeforeEach(func() {
			templatesRepo = mocks.NewTemplatesRepo()
			conn = mocks.NewConnection()
			database = mocks.NewDatabase()
			database.ConnectionCall.Returns.Connection = conn

			finder = services.NewTemplateFinder(templatesRepo)
		})

		Context("when the finder returns a template", func() {
			Context("when the template exists in the database", func() {
				var expectedTemplate models.Template

				BeforeEach(func() {
					expectedTemplate = models.Template{
						ID:      "awesome-template-id",
						Name:    "Awesome New Template",
						Subject: "Wow this is really awesome",
						Text:    "awesome new hungry raptors template",
						HTML:    "<p>hungry raptors are newly awesome template</p>",
					}
					templatesRepo.Templates["awesome-template-id"] = expectedTemplate
				})

				It("returns the requested template", func() {
					template, err := finder.FindByID(database, "awesome-template-id")
					Expect(err).ToNot(HaveOccurred())

					Expect(template).To(Equal(expectedTemplate))
					Expect(templatesRepo.FindByIDCall.Receives.Connection).To(Equal(conn))
					Expect(templatesRepo.FindByIDCall.Receives.TemplateID).To(Equal("awesome-template-id"))
				})
			})

		})

		Context("the finder has an error", func() {
			It("propagates the error", func() {
				templatesRepo.FindError = errors.New("some-error")
				templatesRepo.Templates["some-template-id"] = models.Template{
					Name:    "Not nice template",
					Subject: "Not the kind you want",
					Text:    "throwing errors template",
					HTML:    "<h1>Wow you are a throwing errors!</h1>",
				}
				_, err := finder.FindByID(database, "some-template-id")
				Expect(err.Error()).To(Equal("some-error"))
			})
		})
	})
})
