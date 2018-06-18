package payment

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
	g := e.Group("/payment")
	g.POST("/charge", createChargeHandler)
}

func createChargeHandler(c echo.Context) error {
	if client == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	redirect := c.QueryParam("redirect")
	token := &tokenRequest{}
	if redirect != "" {
		id := c.FormValue("stripeToken")
		token.StripeToken.ID = id
	} else {
		if err := c.Bind(token); err != nil {
			c.Logger().Errorf("json error: %+v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
		}
	}

	url := envy.Get("PAYMENT_DEMO_LAMBDA", "paymentdemobagws")
	body, _ := json.Marshal(token)
	c.Logger().Debugf("token: %+v", token)
	req := &lambda.InvokeInput{
		FunctionName: &url,
		Payload:      body,
	}

	val, err := client.Invoke(req)
	if err != nil {
		c.Logger().Errorf("Error invoking lambda with error: %+v", err)
		if redirect != "" {
			return c.Redirect(302, redirect+"?error=internal&status=500")
		}
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
			c.Logger().Errorf("json error: %+v", err)
			if redirect != "" {
				return c.Redirect(302, redirect+"?error=internal&status=500")
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var errMsg struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		}
		err = json.Unmarshal([]byte(errResp.ErrorMessage), &errMsg)
		if err != nil {
			c.Logger().Errorf("json error: %+v", err)
			if redirect != "" {
				return c.Redirect(302, redirect+"?error=internal&status=500")
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		c.Logger().Infof("Error creating charge: %+v", string(payload))
		if status := errMsg.Status; status >= 400 {
			if redirect != "" {
				return c.Redirect(302, redirect+"?error="+errMsg.Message+"&status="+strconv.Itoa(status))
			}
			return c.JSONBlob(status, payload)
		}

		if redirect != "" {
			return c.Redirect(302, redirect+"?error="+errMsg.Message+"&status=500")
		}
		return c.JSONBlob(500, payload)
	}

	if redirect != "" {
		return c.Redirect(302, redirect)
	}
	return c.JSONBlob(201, payload)
}
