package jwt

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockInfo = JWTInfo{
		UID:   10000,
		Roles: "user,admin",
		IP:    "192.168.128.1",
	}
)

func TestNewTokenHMAC(t *testing.T) {
	Convey("Given some mock mockInfo sampleSecretKey", t, func() {
		Convey("should return a token", func() {
			token, err := NewTokenHMAC(mockInfo, 7200)
			So(err, ShouldBeNil)
			So(token, ShouldNotBeBlank)
		})
		Convey("generate another token", func() {
			token, err := NewTokenHMAC(mockInfo, 7200)
			fakeMockInfo := mockInfo
			secondToken, err := NewTokenHMAC(fakeMockInfo, 7200)
			So(err, ShouldBeNil)
			So(token, ShouldNotBeBlank)
			So(secondToken, ShouldNotEqual, token)
		})
	})
}

func TestTokenValidate(t *testing.T) {
	Convey("Generate a token expired in 7200 seconds", t, func() {
		token, err := NewTokenHMAC(mockInfo, 7200)
		So(err, ShouldBeNil)
		So(token, ShouldNotBeBlank)
		Convey("TokenValidate should return mockInfo without error", func() {
			parsedInfo, err := TokenValidate(token)
			So(err, ShouldBeNil)
			So(parsedInfo.Expires, ShouldNotEqual, mockInfo.Expires)
		})
	})
	Convey("Generate a token expired in 0 second", t, func() {
		token, _ := NewTokenHMAC(mockInfo, 0)
		Convey("TokenValidate should be error due to expired", func() {
			time.Sleep(time.Second * 1)
			mockInfo, err := TokenValidate(token)
			So(err, ShouldNotBeNil)
			So(mockInfo, ShouldBeNil)
		})
	})
	Convey("Validate different mockInfo", t, func() {
		token, _ := NewTokenHMAC(mockInfo, 7200)
		fakeMockInfo := mockInfo
		anothertoken, _ := NewTokenHMAC(fakeMockInfo, 7200)
		Convey("TokenValidate should parse different mockInfo", func() {
			mockInfo, err := TokenValidate(token)
			mockInfo1, err := TokenValidate(anothertoken)
			So(err, ShouldBeNil)
			So(mockInfo, ShouldNotBeNil)
			So(mockInfo1, ShouldNotBeNil)
		})
	})
}
