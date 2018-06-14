package payment

import (
	"encoding/json"
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
		log.Printf("Error creating charge with error: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create charge", err)
	}

	return c.JSONBlob(201, val.Payload)
}
