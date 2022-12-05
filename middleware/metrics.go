package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	ginParamRegexp = regexp.MustCompile(`:[^/\\]*`)
	ginAnyRegexp   = regexp.MustCompile(`\*.*`)
	slashRegexp    = regexp.MustCompile(`/`)
)

const other string = "other"

func urlLabelMapping(c *gin.Context, endpoints []*endpoint, metricsPath string) (path, method string) {
	requestPath := c.Request.URL.Path
	requestMethod := c.Request.Method
	if requestPath == metricsPath && requestMethod == http.MethodGet {
		return requestPath, requestMethod
	}

	for _, r := range endpoints {
		if r.regexp.MatchString(requestPath) && requestMethod == r.method {
			return r.templatePath, r.method
		}
	}
	return other, other
}

func GinPathToRegexp(fullPath string) *regexp.Regexp {
	regexPath := slashRegexp.ReplaceAllString(fullPath, `\/`)
	regexPath = ginParamRegexp.ReplaceAllString(regexPath, `[^\/]+`)
	regexPath = ginAnyRegexp.ReplaceAllString(regexPath, `.*`)
	return regexp.MustCompile("^" + regexPath + `$`)
}

type endpoint struct {
	regexp       *regexp.Regexp
	templatePath string
	method       string
}

type Prometheus struct {
	reqCnt               *prometheus.CounterVec
	reqDur, reqSz, resSz prometheus.Summary
	metricsPath          string
	endpoints            []*endpoint
}

func (p *Prometheus) RegisteredEndpointsCount() int {
	return len(p.endpoints)
}

func (p *Prometheus) Add(path, method string) {
	p.endpoints = append(p.endpoints, &endpoint{
		templatePath: path,
		method:       method,
		regexp:       GinPathToRegexp(path),
	})
}

func NewPrometheus(metricsPath, subsystem string, log logrus.FieldLogger) *Prometheus {
	p := &Prometheus{
		metricsPath: metricsPath,
	}
	p.registerMetrics(subsystem, log)

	return p
}

func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.handlerFunc())
	e.GET(p.metricsPath, prometheusHandler())
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func (p *Prometheus) registerMetrics(subsystem string, log logrus.FieldLogger) {
	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "service_version", "host", "url"},
	)
	if err := prometheus.Register(p.reqCnt); err != nil {
		log.WithError(err).Error("requests_total could not be registered in Prometheus")
	}

	p.reqDur = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
		},
	)
	if err := prometheus.Register(p.reqDur); err != nil {
		log.WithError(err).Error("request_duration_seconds could not be registered in Prometheus")
	}

	p.resSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)
	if err := prometheus.Register(p.resSz); err != nil {
		log.WithError(err).Error("response_size_bytes could not be registered in Prometheus")
	}

	p.reqSz = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)
	if err := prometheus.Register(p.reqSz); err != nil {
		log.WithError(err).Error("request_size_bytes could not be registered in Prometheus")
	}
}

func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

func (p *Prometheus) handlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		url, method := urlLabelMapping(c, p.endpoints, p.metricsPath)

		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSz := float64(c.Writer.Size())

		p.reqDur.Observe(elapsed)
		p.reqCnt.WithLabelValues(status, method, viper.GetString("service_version"), c.Request.Host, url).Inc()
		p.reqSz.Observe(float64(reqSz))
		p.resSz.Observe(resSz)
	}
}
