package apps

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func (a *App) GET(path string) Payload {
	return a.GETf("%s", path)
}

func (a *App) GETf(format string, s ...any) Payload {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusOK))

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())

	GinkgoWriter.Printf("Recieved: %s\n", string(data))
	return Payload(data)
}

func (a *App) GETResponse(path string) *http.Response {
	return a.GETResponsef("%s", path)
}

// GETResponsef does an HTTP get, returning the *http.Response
func (a *App) GETResponsef(format string, s ...any) *http.Response {
	GinkgoHelper()

	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP GET: %s\n", url)
	response, err := http.Get(url)
	Expect(err).NotTo(HaveOccurred())
	return response
}

func (a *App) PUT(data, path string) {
	a.PUTf(data, "%s", path)
}

func (a *App) PUTf(data, format string, s ...any) {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP PUT: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPut, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "text/html")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))
}

func (a *App) POST(data, path string) Payload {
	return a.POSTf(data, "%s", path)
}

func (a *App) POSTf(data, format string, s ...any) Payload {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP POST: %s\n", url)
	GinkgoWriter.Printf("Sending data: %s\n", data)
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data))
	Expect(err).NotTo(HaveOccurred())
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusCreated, http.StatusOK))

	responseBody, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	Expect(err).NotTo(HaveOccurred())
	return Payload(responseBody)
}

func (a *App) DELETETestTable() {
	url := a.urlf("/")
	GinkgoWriter.Printf("HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent))
}

func (a *App) DELETE(path string) {
	a.DELETEf("%s", path)
}

func (a *App) DELETEf(format string, s ...any) {
	url := a.urlf(format, s...)
	GinkgoWriter.Printf("HTTP DELETE: %s\n", url)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	Expect(err).NotTo(HaveOccurred())

	response, err := http.DefaultClient.Do(request)
	Expect(err).NotTo(HaveOccurred())
	Expect(response).To(HaveHTTPStatus(http.StatusGone, http.StatusNoContent, http.StatusOK))
}

func (a *App) urlf(format string, s ...any) string {
	base := a.URL
	path := fmt.Sprintf(format, s...)
	switch {
	case len(path) == 0:
		return base
	case path[0] != '/':
		return fmt.Sprintf("%s/%s", base, path)
	default:
		return base + path
	}
}
