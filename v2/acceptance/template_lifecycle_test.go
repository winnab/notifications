package acceptance

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry-incubator/notifications/v2/acceptance/support"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type templateResponse struct {
	ID       string
	Name     string
	Text     string
	HTML     string
	Subject  string
	Metadata map[string]string
	Links    struct {
		Self struct {
			Href string
		}
	} `json:"_links"`
}
type templatesListResponse struct {
	Templates []templateResponse
	Links     struct {
		Self struct {
			Href string
		}
	} `json:"_links"`
}

var _ = Describe("Template lifecycle", func() {
	var (
		client     *support.Client
		token      string
		templateID string
	)

	BeforeEach(func() {
		client = support.NewClient(support.Config{
			Host:              Servers.Notifications.URL(),
			Trace:             Trace,
			RoundTripRecorder: roundtripRecorder,
		})
		var err error
		token, err = GetClientTokenWithScopes("notifications.write")
		Expect(err).NotTo(HaveOccurred())
	})

	It("can create a new template, retrieve, list and delete", func() {
		By("creating a template", func() {
			var response templateResponse

			client.Document("template-create")
			status, err := client.DoTyped("POST", "/templates", map[string]interface{}{
				"name":    "An interesting template",
				"text":    "template text",
				"html":    "template html",
				"subject": "template subject",
				"metadata": map[string]interface{}{
					"template": "metadata",
				},
			}, token, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusCreated))

			templateID = response.ID

			Expect(templateID).NotTo(BeEmpty())
			Expect(response.Name).To(Equal("An interesting template"))
			Expect(response.Text).To(Equal("template text"))
			Expect(response.HTML).To(Equal("template html"))
			Expect(response.Subject).To(Equal("template subject"))
			Expect(response.Metadata["template"]).To(Equal("metadata"))
			Expect(response.Links.Self.Href).To(Equal(fmt.Sprintf("/templates/%s", templateID)))
		})

		By("getting the template", func() {
			var response templateResponse

			client.Document("template-get")
			status, err := client.DoTyped("GET", fmt.Sprintf("/templates/%s", templateID), nil, token, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))

			Expect(response.ID).To(Equal(templateID))
			Expect(response.Name).To(Equal("An interesting template"))
			Expect(response.Text).To(Equal("template text"))
			Expect(response.HTML).To(Equal("template html"))
			Expect(response.Subject).To(Equal("template subject"))
			Expect(response.Metadata["template"]).To(Equal("metadata"))
			Expect(response.Links.Self.Href).To(Equal(fmt.Sprintf("/templates/%s", templateID)))
		})

		By("updating the template", func() {
			var response templateResponse
			url := fmt.Sprintf("/templates/%s", templateID)
			client.Document("template-update")
			status, err := client.DoTyped("PUT", url, map[string]interface{}{
				"name":    "A more interesting template",
				"text":    "text",
				"html":    "html",
				"subject": "subject",
				"metadata": map[string]interface{}{
					"banana": "something",
				},
			}, token, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))

			Expect(response.ID).To(Equal(templateID))
			Expect(response.Name).To(Equal("A more interesting template"))
			Expect(response.Text).To(Equal("text"))
			Expect(response.HTML).To(Equal("html"))
			Expect(response.Subject).To(Equal("subject"))
			Expect(response.Metadata["banana"]).To(Equal("something"))
			Expect(response.Links.Self.Href).To(Equal(url))
		})

		By("getting the template", func() {
			status, response, err := client.Do("GET", fmt.Sprintf("/templates/%s", templateID), nil, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))

			Expect(response["id"]).To(Equal(templateID))
			Expect(response["name"]).To(Equal("A more interesting template"))
			Expect(response["text"]).To(Equal("text"))
			Expect(response["html"]).To(Equal("html"))
			Expect(response["subject"]).To(Equal("subject"))
			Expect(response["metadata"]).To(Equal(map[string]interface{}{
				"banana": "something",
			}))
		})

		By("creating a template for another client", func() {
			otherClientToken, err := GetClientTokenWithScopes("notifications.write")
			Expect(err).NotTo(HaveOccurred())

			status, _, err := client.Do("POST", "/templates", map[string]interface{}{
				"name":    "An invisible template",
				"text":    "template text",
				"html":    "template html",
				"subject": "template subject",
				"metadata": map[string]interface{}{
					"template": "metadata",
				},
			}, otherClientToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusCreated))
		})

		By("listing all templates", func() {
			var response templatesListResponse

			client.Document("template-list")
			status, err := client.DoTyped("GET", "/templates", nil, token, &response)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))

			Expect(response.Templates).To(HaveLen(1))
			Expect(response.Templates[0].ID).To(Equal(templateID))
			Expect(response.Templates[0].Links.Self.Href).To(Equal(fmt.Sprintf("/templates/%s", templateID)))

			Expect(response.Links.Self.Href).To(Equal("/templates"))
		})

		By("deleting the template", func() {
			client.Document("template-delete")
			status, _, err := client.Do("DELETE", fmt.Sprintf("/templates/%s", templateID), nil, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusNoContent))
		})

		By("failing to get the deleted template", func() {
			status, _, err := client.Do("GET", fmt.Sprintf("/templates/%s", templateID), nil, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(http.StatusNotFound))
		})
	})

	Context("when omitting field values", func() {
		It("uses the existing value", func() {
			By("creating a template", func() {
				status, response, err := client.Do("POST", "/templates", map[string]interface{}{
					"name":    "An interesting template",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusCreated))

				templateID = response["id"].(string)

				Expect(templateID).NotTo(BeEmpty())
				Expect(response["name"]).To(Equal("An interesting template"))
				Expect(response["text"]).To(Equal("template text"))
				Expect(response["html"]).To(Equal("template html"))
				Expect(response["subject"]).To(Equal("template subject"))
				Expect(response["metadata"]).To(Equal(map[string]interface{}{
					"template": "metadata",
				}))
			})

			By("updating the template", func() {
				status, response, err := client.Do("PUT", fmt.Sprintf("/templates/%s", templateID), map[string]interface{}{}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusOK))

				Expect(response["id"]).To(Equal(templateID))
				Expect(response["name"]).To(Equal("An interesting template"))
				Expect(response["text"]).To(Equal("template text"))
				Expect(response["html"]).To(Equal("template html"))
				Expect(response["subject"]).To(Equal("template subject"))
				Expect(response["metadata"]).To(Equal(map[string]interface{}{
					"template": "metadata",
				}))
			})
		})
	})
	Context("when clearing field values", func() {
		It("sets them back to their default values", func() {
			By("creating a template", func() {
				status, response, err := client.Do("POST", "/templates", map[string]interface{}{
					"name":    "An interesting template",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusCreated))

				templateID = response["id"].(string)

				Expect(templateID).NotTo(BeEmpty())
				Expect(response["name"]).To(Equal("An interesting template"))
				Expect(response["text"]).To(Equal("template text"))
				Expect(response["html"]).To(Equal("template html"))
				Expect(response["subject"]).To(Equal("template subject"))
				Expect(response["metadata"]).To(Equal(map[string]interface{}{
					"template": "metadata",
				}))
			})

			By("updating the template", func() {
				status, response, err := client.Do("PUT", fmt.Sprintf("/templates/%s", templateID), map[string]interface{}{
					"name":    "A more interesting template",
					"text":    "text",
					"html":    "html",
					"subject": "",
					"metadata": map[string]interface{}{
						"banana": "something",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusOK))

				Expect(response["id"]).To(Equal(templateID))
				Expect(response["name"]).To(Equal("A more interesting template"))
				Expect(response["text"]).To(Equal("text"))
				Expect(response["html"]).To(Equal("html"))
				Expect(response["subject"]).To(Equal("{{.Subject}}"))
				Expect(response["metadata"]).To(Equal(map[string]interface{}{
					"banana": "something",
				}))
			})
		})
	})

	Context("failure states", func() {
		Context("creating", func() {
			It("returns a 422 when the name field is empty", func() {
				status, response, err := client.Do("POST", "/templates", map[string]interface{}{
					"name":    "",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(422))
				Expect(response["errors"]).To(ContainElement("Template \"name\" field cannot be empty"))
			})
		})

		Context("updating", func() {
			It("returns a 404 when the template ID does not exist", func() {
				status, response, err := client.Do("PUT", "/templates/bogus", map[string]interface{}{
					"name":    "An interesting template",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusNotFound))
				Expect(response["errors"]).To(ContainElement("Template with id \"bogus\" could not be found"))
			})

			It("returns a 422 when the name field is empty", func() {
				status, response, err := client.Do("POST", "/templates", map[string]interface{}{
					"name":    "An interesting template",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusCreated))

				status, response, err = client.Do("PUT", fmt.Sprintf("/templates/%s", response["id"]), map[string]interface{}{
					"name":    "",
					"text":    "template text",
					"html":    "template html",
					"subject": "template subject",
					"metadata": map[string]interface{}{
						"template": "metadata",
					},
				}, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(422))
				Expect(response["errors"]).To(ContainElement("Template \"name\" field cannot be empty"))
			})

			It("returns a 422 when text and html would be empty", func() {
				By("creating a template", func() {
					status, response, err := client.Do("POST", "/templates", map[string]interface{}{
						"name":    "An interesting template",
						"html":    "template html",
						"subject": "template subject",
						"metadata": map[string]interface{}{
							"template": "metadata",
						},
					}, token)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(http.StatusCreated))

					templateID = response["id"].(string)
				})

				By("updating the template", func() {
					status, response, err := client.Do("PUT", fmt.Sprintf("/templates/%s", templateID), map[string]interface{}{
						"html": "",
					}, token)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(422))
					Expect(response["errors"]).To(ContainElement("missing either template text or html"))
				})
			})
		})

		Context("getting", func() {
			It("returns a 404 when the template cannot be retrieved", func() {
				status, response, err := client.Do("GET", "/templates/missing-template-id", nil, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusNotFound))
				Expect(response["errors"]).To(ContainElement("Template with id \"missing-template-id\" could not be found"))
			})

			It("returns a 404 when the template belongs to a different client", func() {
				var templateID string

				By("creating a template for one client", func() {
					status, response, err := client.Do("POST", "/templates", map[string]interface{}{
						"name":    "An interesting template",
						"text":    "template text",
						"html":    "template html",
						"subject": "template subject",
						"metadata": map[string]interface{}{
							"template": "metadata",
						},
					}, token)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(http.StatusCreated))

					templateID = response["id"].(string)
				})

				By("attempting to access the created template as another client", func() {
					otherClientToken, err := GetClientTokenWithScopes("notifications.write")
					Expect(err).NotTo(HaveOccurred())

					status, response, err := client.Do("GET", fmt.Sprintf("/templates/%s", templateID), nil, otherClientToken)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(http.StatusNotFound))
					Expect(response["errors"]).To(ContainElement(fmt.Sprintf("Template with id %q could not be found", templateID)))
				})
			})
		})

		Context("deleting", func() {
			It("returns a 404 when the template to delete does not exist", func() {
				status, response, err := client.Do("DELETE", "/templates/missing-template-id", nil, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusNotFound))
				Expect(response["errors"]).To(ContainElement("Template with id \"missing-template-id\" could not be found"))
			})
		})
	})

	Context("when interacting with the default template", func() {
		Context("getting", func() {
			It("returns the default template with the default values when it has never been set before", func() {

				var response struct {
					ID       string
					Name     string
					Text     string
					HTML     string
					Subject  string
					Metadata map[string]string
					Links    struct {
						Self struct {
							Href string
						}
					} `json:"_links"`
				}
				status, err := client.DoTyped("GET", "/templates/default", nil, token, &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).To(Equal(http.StatusOK))

				Expect(response.ID).To(Equal("default"))
				Expect(response.Name).To(Equal("The Default Template"))
				Expect(response.Text).To(Equal("{{.Text}}"))
				Expect(response.HTML).To(Equal("{{.HTML}}"))
				Expect(response.Subject).To(Equal("{{.Subject}}"))
				Expect(response.Links.Self.Href).To(Equal("/templates/default"))
			})
		})

		Context("updating", func() {
			var adminToken string

			BeforeEach(func() {
				var err error
				adminToken, err = GetClientTokenWithScopes("notifications.admin")
				Expect(err).NotTo(HaveOccurred())
			})

			It("persists the updated default template", func() {
				By("updating the default template", func() {
					var response struct {
						ID       string
						Name     string
						Text     string
						HTML     string
						Subject  string
						Metadata map[string]string
						Links    struct {
							Self struct {
								Href string
							}
						} `json:"_links"`
					}

					url := "/templates/default"
					status, err := client.DoTyped("PUT", url, map[string]interface{}{
						"name":     "some other default",
						"text":     "in default",
						"html":     "massively defaulting",
						"subject":  "time to default",
						"metadata": map[string]interface{}{},
					}, adminToken, &response)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(http.StatusOK))

					Expect(response.ID).To(Equal("default"))
					Expect(response.Name).To(Equal("some other default"))
					Expect(response.Text).To(Equal("in default"))
					Expect(response.HTML).To(Equal("massively defaulting"))
					Expect(response.Subject).To(Equal("time to default"))
					Expect(response.Links.Self.Href).To(Equal(url))

				})

				By("retrieving the newly updated default template", func() {
					status, response, err := client.Do("GET", "/templates/default", nil, token)
					Expect(err).NotTo(HaveOccurred())
					Expect(status).To(Equal(http.StatusOK))

					Expect(response["id"]).To(Equal("default"))
					Expect(response["name"]).To(Equal("some other default"))
					Expect(response["text"]).To(Equal("in default"))
					Expect(response["html"]).To(Equal("massively defaulting"))
					Expect(response["subject"]).To(Equal("time to default"))
					Expect(response["metadata"]).To(Equal(map[string]interface{}{}))
				})
			})
		})
	})
})
