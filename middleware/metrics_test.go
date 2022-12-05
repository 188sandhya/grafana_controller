// +build unitTests

package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

const statusOkPath = "/status/ok"
const statusNotFoundPath = "/status/not_found"

const paramPath = "/some/:path/with/:param"
const paramPathExampleValue = "/some/param_val_1/with/param_val_2"

const anyParamPath = "/anyparam/test/*endpoint"
const anyParamPathExample = "/anyparam/test/val1/val2/val3"

var _ = Describe("Prometheus Metrics Middleware", func() {

	Describe("GinPathToRegexp()", func() {
		Context("When parametrized path added", func() {
			rAnyParam := middleware.GinPathToRegexp(anyParamPath)
			It("returns proper regexp", func() {
				Expect(rAnyParam.String()).To(Equal(`^\/anyparam\/test\/.*$`))
				Expect(rAnyParam.MatchString(anyParamPathExample)).To(BeTrue())
			})

			rParam := middleware.GinPathToRegexp(paramPath)
			It("returns proper regexp", func() {
				Expect(rParam.String()).To(Equal(`^\/some\/[^\/]+\/with\/[^\/]+$`))
				Expect(rParam.MatchString(paramPathExampleValue)).To(BeTrue())
			})

			r := middleware.GinPathToRegexp(statusOkPath)
			It("returns proper regexp", func() {
				Expect(r.String()).To(Equal(`^\/status\/ok$`))
				Expect(r.MatchString(statusOkPath)).To(BeTrue())
			})
		})
	})

	Describe("Add()", func() {
		logger := logrus.New()
		prom := middleware.NewPrometheus("metricsPath", "subsystem", logger)
		Context("When endpoint added", func() {
			prom.Add("/my/test/path", http.MethodPost)
			prom.Add("/my/second/path", http.MethodPost)
			It("appears on the list", func() {
				Expect(prom.RegisteredEndpointsCount()).To(Equal(2))
			})
		})
	})

	Describe("Use()", func() {
		var ginEngine *gin.Engine
		var prom *middleware.Prometheus

		var responseRecorder *httptest.ResponseRecorder
		var metricsResponseRecorder *httptest.ResponseRecorder

		statusOkHandler := func(context *gin.Context) {
			context.Status(http.StatusOK)
		}
		statusNotFoundHandler := func(context *gin.Context) {
			context.Status(http.StatusNotFound)
		}

		BeforeEach(func() {
			gin.SetMode(gin.TestMode)
			ginEngine = gin.New()
			logger := logrus.New()

			prom = middleware.NewPrometheus("/metrics", "test", logger)
			prom.Add(statusOkPath, http.MethodGet)
			prom.Add(statusNotFoundPath, http.MethodGet)
			prom.Add(paramPath, http.MethodGet)

			prom.Use(ginEngine)

			ginEngine.GET(statusOkPath, statusOkHandler)
			ginEngine.GET(statusNotFoundPath, statusNotFoundHandler)
			ginEngine.GET(paramPath, statusOkHandler)

			metricsResponseRecorder = httptest.NewRecorder()
		})

		Context("", func() {

			Context("when some requests were executed", func() {
				statusOkRequestsCount := 2
				BeforeEach(func() {
					for i := 0; i < statusOkRequestsCount; i++ {
						req, _ := http.NewRequest(http.MethodGet, statusOkPath, nil)
						responseRecorder = httptest.NewRecorder()
						ginEngine.ServeHTTP(responseRecorder, req)
						Expect(responseRecorder.Code).To(Equal(http.StatusOK))
					}

					req, _ := http.NewRequest(http.MethodGet, statusNotFoundPath, nil)
					responseRecorder = httptest.NewRecorder()
					ginEngine.ServeHTTP(responseRecorder, req)
					Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))

					req2, _ := http.NewRequest(http.MethodGet, "/some/other/path", nil)
					responseRecorder = httptest.NewRecorder()
					ginEngine.ServeHTTP(responseRecorder, req2)
					Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))

					metricsReq, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
					ginEngine.ServeHTTP(metricsResponseRecorder, metricsReq)
				})
				It("should be available at metrics endpoint", func() {
					Expect(metricsResponseRecorder.Code).To(Equal(http.StatusOK))
					Expect(metricsResponseRecorder.Body.String()).To(ContainSubstring(`test_requests_total{code="%d",host="",method="GET",service_version="",url="%s"} %d`, http.StatusOK, statusOkPath, statusOkRequestsCount))
					Expect(metricsResponseRecorder.Body.String()).To(ContainSubstring(`test_requests_total{code="%d",host="",method="GET",service_version="",url="%s"} %d`, http.StatusNotFound, statusNotFoundPath, 1))
					Expect(metricsResponseRecorder.Body.String()).To(ContainSubstring(`test_requests_total{code="%d",host="",method="%s",service_version="",url="%s"} %d`, http.StatusNotFound, "other", "other", 1))
				})
			})
		})
	})
})
