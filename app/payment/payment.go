package payment

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/gobuffalo/envy"
	"github.com/labstack/echo"
	stripe "github.com/stripe/stripe-go"
)

var client *lambda.Lambda

type tokenRequest struct {
	StripeToken *stripe.Token `json:"stripeToken"`
}

func init() {
	sess, err := session.NewSession()
	if err != nil {
		log.Printf("Unable to create session: %+v", err)
	}

	client = lambda.New(sess, aws.NewConfig().WithRegion("us-east-1"))
}

// Register creates an Echo group for payment
func Register(e *echo.Echo) {
	log.Println("Registering payment group")
	g := e.Group("/payment")
	g.POST("/charge", createChargeHandler)
}

func createChargeHandler(c echo.Context) error {
	c.Logger().Infof("pre: %d", time.Now().Unix())
	if client == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	token := &tokenRequest{}
	if err := c.Bind(token); err != nil {
		c.Logger().Errorf("json error: %+v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}
	c.Logger().Infof("after bind: %d", time.Now().Unix())

	url := envy.Get("PAYMENT_DEMO_LAMBDA", "paymentdemobagws")
	body, _ := json.Marshal(token)
	c.Logger().Infof("token: %+v", token)
	req := &lambda.InvokeInput{
		FunctionName: &url,
		Payload:      body,
	}
	c.Logger().Infof("before invoke: %d", time.Now().Unix())

	val, err := client.Invoke(req)
	if err != nil {
		c.Logger().Errorf("Error invoking lambda with error: %+v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create charge", err)
	}
	c.Logger().Infof("after invoke: %d", time.Now().Unix())

	payload := val.Payload

	if val.FunctionError != nil {
		var errResp struct {
			ErrorMessage string `json:"errorMessage"`
			ErrorType    string `json:"errorType"`
		}

		err := json.Unmarshal(payload, &errResp)
		if err != nil {
			c.Logger().Errorf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var errMsg struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}
		err = json.Unmarshal([]byte(errResp.ErrorMessage), &errMsg)
		if err != nil {
			c.Logger().Errorf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		c.Logger().Infof("Error creating charge: %+v", string(payload))
		if status := errMsg.Status; status >= 400 {
			return c.JSONBlob(status, payload)
		}

		return c.JSONBlob(500, payload)
	}

	return c.JSONBlob(201, payload)
}
