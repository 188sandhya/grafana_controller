// +build unitTests

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("helpers", func() {

	var contextMock *gin.Context
	var params gin.Params
	var userContext *auth.UserContext

	JustBeforeEach(func() {
		contextMock = &gin.Context{
			Params: params,
		}
		contextMock.Set("UserContext", userContext)
	})
	Describe("GetIDParam()", func() {
		Context("when gin.Context contains param 'id' with numeric value", func() {
			BeforeEach(func() {
				params = []gin.Param{
					{
						Key:   "id",
						Value: "10",
					},
				}
			})
			It("returns 'id' param as int64", func() {
				id, err := GetIDParam(contextMock)
				Expect(err).ToNot(HaveOccurred())
				Expect(id).To(Equal(int64(10)))
			})
		})

		Context("when gin.Context contains param 'id' with not numeric value", func() {
			BeforeEach(func() {
				params = []gin.Param{
					{
						Key:   "id",
						Value: "abc10",
					},
				}
			})
			It("returns correct error message", func() {
				_, err := GetIDParam(contextMock)
				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.ParseErrors)).To(BeTrue())
			})
		})

		Context("when gin.Context contains param 'id' with wrong number", func() {
			BeforeEach(func() {
				params = []gin.Param{
					{
						Key:   "id",
						Value: "-1",
					},
				}
			})
			It("returns correct error message", func() {
				_, err := GetIDParam(contextMock)
				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.ValidationErrors)).To(BeTrue())
			})
		})

		Context("when gin.Context contains param 'id' with wrong number", func() {
			BeforeEach(func() {
				params = []gin.Param{
					{
						Key:   "id",
						Value: "9999999999",
					},
				}
			})
			It("returns correct error message", func() {
				_, err := GetIDParam(contextMock)
				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.ValidationErrors)).To(BeTrue())
			})
		})

		Context("when gin.Context doesn't contain param 'id'", func() {
			BeforeEach(func() {
				params = []gin.Param{}
			})
			It("returns correct error message", func() {
				_, err := GetIDParam(contextMock)
				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.ParseErrors)).To(BeTrue())
			})
		})
	})

	Describe("GetUserContext()", func() {
		Context("when gin.Context contains UserContext with set values", func() {
			BeforeEach(func() {
				userContext = &auth.UserContext{ID: 5, Cookie: "test cookie"}
			})
			It("returns 'id' param as int64", func() {
				userContext, err := GetUserContext(contextMock)
				Expect(err).ToNot(HaveOccurred())
				Expect(userContext.ID).To(Equal(int64(5)))
				Expect(userContext.Cookie).To(Equal("test cookie"))
			})
		})
	})
})
