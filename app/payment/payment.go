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

	if val.FunctionError != nil {
		var errResp struct {
			ErrorMessage string `json:"errorMessage"`
			ErrorType    string `json:"errorType"`
		}

		err := json.Unmarshal(payload, &errResp)
		if err != nil {
			fmt.Printf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var errMsg struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}
		err = json.Unmarshal([]byte(errResp.ErrorMessage), &errMsg)
		if err != nil {
			fmt.Printf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		log.Printf("Error creating charge: %+v", payload)
		if status := errMsg.Status; status >= 400 {
			return c.JSONBlob(status, payload)
		}

		return c.JSONBlob(500, payload)
	}

	return c.JSONBlob(201, payload)
}
