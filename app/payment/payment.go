package payment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/gobuffalo/envy"
	"github.com/labstack/echo"
	stripe "github.com/stripe/stripe-go"
)

var client *lambda.Lambda

func init() {
	sess, err := session.NewSession()
	if err != nil {
		log.Printf("Unable to create session: %+v", err)
	}

	// Config Lambda lives in us-east-1
	client = lambda.New(sess, aws.NewConfig().WithRegion("us-east-1"))
}

// Register creates an Echo group for payment
func Register(e *echo.Echo) {
	g := e.Group("/payment")
	g.POST("/charge", createChargeHandler)
}

func createChargeHandler(c echo.Context) error {
	if client == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	token := &stripe.Token{}
	if err := c.Bind(token); err != nil {
		log.Printf("json error: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	url := envy.Get("PAYMENT_DEMO_LAMBDA", "paymentdemobagws")
	body, _ := json.Marshal(token)
	req := &lambda.InvokeInput{
		FunctionName: &url,
		Payload:      body,
	}

	val, err := client.Invoke(req)
	if err != nil {
		log.Printf("Error invoking lambda with error: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create charge", err)
	}

	payload := val.Payload

	if val.FunctionError == aws.String("Handled") {
		var errResp struct {
			ErrorMessage struct {
				Code    string `json:"code"`
				Status  int    `json:"status"`
				Message string `json:"message"`
				Param   string `json:"param"`
				Type    string `json:"type"`
			} `json:"errorMessage"`
			ErrorType string `json:"errorType"`
		}

		err := json.Unmarshal(payload, &errResp)
		if err != nil {
			fmt.Printf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		log.Printf("Error creating charge: %+v", payload)
		if status := errResp.ErrorMessage.Status; status != 0 {
			return c.JSONBlob(status, payload)
		}

		return c.JSONBlob(500, payload)
	} else if val.FunctionError == aws.String("Unhandled") {
		log.Printf("Unhandled error creating charge: %+v", payload)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSONBlob(201, payload)
}
