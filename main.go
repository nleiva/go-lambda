package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mssola/user_agent"
	_ "github.com/nleiva/go-lambda/statik"
	"github.com/rakyll/statik/fs"
)

type data struct {
	IP       string `json:"This is your IP address"`
	Country  string `json:"You are visiting us from"`
	Platf    string `json:"Platform"`
	OS       string `json:"OS"`
	Browser  string `json:"Browser"`
	Bversion string `json:"Browser Version"`
	Host     string `json:"Target host"`
	Mob      bool   `json:"Mobile"`
	Bot      bool   `json:"Bot"`
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Processing request data for request %s.\n", request.RequestContext.RequestID)
	// fmt.Printf("Body size = %d.\n", len(request.Body))

	statikFS, err := fs.New()
	if err != nil {
		fmt.Println(err.Error())
	}

	// Access individual files by their paths.
	r, err := statikFS.Open("/test.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(contents))

	var ip string
	if val, ok := request.Headers["X-Forwarded-For"]; ok {
		s := strings.Split(val, ",")
		ip = s[0]
	}

	sip := request.RequestContext.Identity.SourceIP

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

	var c string
	if val, ok := request.Headers["CloudFront-Viewer-Country"]; ok {
		c = val
	}

	d := data{
		IP:       ip + " and " + sip,
		Country:  c,
		Platf:    ua.Platform(),
		OS:       ua.OS(),
		Browser:  br,
		Bversion: ver,
		Mob:      ua.Mobile(),
		Bot:      ua.Bot(),
		Host:     h,
	}

	// The APIGatewayProxyResponse.Body field needs to be a string, so
	// we marshal the record into JSON.
	js, err := json.Marshal(d)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       http.StatusText(http.StatusInternalServerError),
		}, nil
	}

	pjs, err := prettyprint(js)
	if err != nil {
		fmt.Printf("error pretty printing: %s\n", string(js))
	}

	return events.APIGatewayProxyResponse{Body: string(pjs),
		StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handleRequest)
}
