package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mssola/user_agent"
	_ "github.com/nleiva/go-lambda/statik"
	"github.com/rakyll/statik/fs"
)

type data struct {
	IP       string `json:"This is your IP address"`
	Proxy    string `json:"AWS proxy"`
	Country  string `json:"You are visiting us from"`
	Platf    string `json:"Platform"`
	OS       string `json:"OS"`
	Browser  string `json:"Browser"`
	Bversion string `json:"Browser Version"`
	Host     string `json:"Target host"`
	Mob      bool   `json:"Mobile"`
	Bot      bool   `json:"Are you a bot"`
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (o events.APIGatewayProxyResponse, e error) {
	fmt.Printf("Processing request data for request %s, from IP %s.\n",
		request.RequestContext.RequestID,
		request.RequestContext.Identity.SourceIP)

	o = events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}

	check := func(err error) bool {
		if err != nil {
			fmt.Println(err.Error())
			o = events.APIGatewayProxyResponse{
				Body: http.StatusText(http.StatusInternalServerError) + ": " + err.Error(),
			}
			return true
		}
		return false
	}

	statikFS, err := fs.New()
	if check(err) {
		return
	}

	// Return a 404 if the requested template doesn't exist. HARDCODED for now.
	fp, err := statikFS.Open(filepath.Join("/templates", "example.html"))
	if check(err) {
		return
	}
	defer fp.Close()

	lp, err := statikFS.Open("/templates/layout.html")
	if check(err) {
		return
	}
	defer lp.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(lp)
	if check(err) {
		return
	}
	lps := buf.String()

	// https://play.golang.org/p/DUkUAHdIGo3
	t, err := template.New("base").Parse(lps)
	if check(err) {
		return
	}

	buf = new(bytes.Buffer)
	_, err = buf.ReadFrom(fp)
	if check(err) {
		return
	}
	fps := buf.String()

	tmpl, err := t.New("layout").Parse(fps)
	if check(err) {
		return
	}

	var ip string
	if val, ok := request.Headers["X-Forwarded-For"]; ok {
		s := strings.Split(val, ",")
		ip = s[0]
	}

	p := "No"
	if val, ok := request.Headers["Via"]; ok {
		p = val
	}

	var sys string
	if val, ok := request.Headers["User-Agent"]; ok {
		sys = val
	}
	ua := user_agent.New(sys)
	br, ver := ua.Browser()

	// Can't forward Host header from CloudFront from API Gateway.
	// It results in {"message":"Forbidden"}.
	// We need something like 'X-Forwarded-Host' instead.
	var h string
	if val, ok := request.Headers["Host"]; ok {
		h = val
	}

	c := "Unknown"
	if val, ok := request.Headers["CloudFront-Viewer-Country"]; ok {
		c = val
	}

	d := data{
		IP:       ip,
		Country:  c,
		Platf:    ua.Platform(),
		OS:       ua.OS(),
		Browser:  br,
		Bversion: ver,
		Mob:      ua.Mobile(),
		Bot:      ua.Bot(),
		Host:     h,
		Proxy:    p,
	}

	var b strings.Builder
	err = tmpl.ExecuteTemplate(&b, "layout", d)
	if check(err) {
		return
	}

	return events.APIGatewayProxyResponse{
		Body: b.String(),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
