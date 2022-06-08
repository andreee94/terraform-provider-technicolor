package technicolor

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

// class
type TechnicolorRouter struct {
	Address   string
	Port      int
	CSRFToken string
	url       string
	sessionID string
	collector *colly.Collector
	user      *CryptoUser
	s         []byte
	B         []byte
	M         []byte
}

func NewTechnicolorRouter(address string, port int) *TechnicolorRouter {
	return &TechnicolorRouter{
		Address: address,
		Port:    port,
		url:     fmt.Sprintf("http://%s:%d", address, port),
		collector: colly.NewCollector(
			colly.AllowURLRevisit(),
			colly.Headers(map[string]string{}),
		),
	}
}

func (router *TechnicolorRouter) getEndpoint(endpoint string) string {
	return fmt.Sprintf("%s/%s", router.url, endpoint)
}

func (router *TechnicolorRouter) Login(username string, password string) (err error, isAuthenticated bool) {

	err = router.getCSRFToken()

	if err != nil {
		return err, false
	}

	router.user = NewCryptoUser(username, password, SHA256, NG_2048)

	I, A := router.user.StartAuthentication()

	// {"CSRFtoken": token, "I": uname, "A": binascii.hexlify(A)}
	err, responseSB := router.authenticate(map[string]string{
		"CSRFtoken": router.CSRFToken,
		"I":         I,
		"A":         hex.EncodeToString(A),
	})

	if err != nil {
		return err, false
	}

	err = router.getSAndB(responseSB)

	if err != nil {
		return err, false
	}

	computedM, _ := router.user.ProcessChallenge(router.s, router.B)

	err, responseChallenge := router.authenticate(map[string]string{
		"CSRFtoken": router.CSRFToken,
		"M":         hex.EncodeToString(computedM),
	})

	err = router.getM(responseChallenge)

	if err != nil {
		return err, false
	}

	isAuthenticated = router.user.ValidateAuthentication(router.M)

	if isAuthenticated {
		router.sessionID = router.collector.Cookies(router.url)[0].Value // TODO check if this is the correct cookie

		// update csrf with new session id
		router.getCSRFTokenAfterLogin()
	}

	// GetCSRFToken
	// Initialize CryptoUser
	// Authenticate
	// Extract s and B for the challange
	// Send challenge response
	// Check that the challenge was successful
	return nil, isAuthenticated
}

func (router *TechnicolorRouter) getCSRFToken() (err error) {
	router.collector.OnHTML("head > meta:nth-child(3)", func(e *colly.HTMLElement) {
		// log.Println("getCSRFToken.OnHTML()")
		router.CSRFToken = e.Attr("content")
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("getCSRFToken.OnError()")
		log.Println("getCSRFToken => error:", err, r.Body)
	})

	err = router.collector.Visit(router.url)

	router.collector.OnHTMLDetach("head > meta:nth-child(3)")

	return err
}

func (router *TechnicolorRouter) getCSRFTokenAfterLogin() (err error) {
	url := router.getEndpoint(TECHNICOLOR_ENDPOINT_LOGIN)
	// url := fmt.Sprintf("%s/login.lp?action=lastaccess", router.url)

	router.collector.OnHTML("head > meta:nth-child(3)", func(e *colly.HTMLElement) {
		// log.Println("getCSRFToken.OnHTML()")
		router.CSRFToken = e.Attr("content")
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("getCSRFToken.OnError()")
		log.Println("getCSRFToken => error:", err, r.Body)
	})

	err = router.collector.Visit(url)

	router.collector.OnHTMLDetach("head > meta:nth-child(3)")

	return err
}

func (router *TechnicolorRouter) authenticate(data map[string]string) (err error, response map[string]string) {
	url := router.getEndpoint(TECHNICOLOR_ENDPOINT_AUTHENTICATE)
	// url := fmt.Sprintf("%s/authenticate", router.url)

	// router.collector.OnHTML("head > meta:nth-child(3)", func(e *colly.HTMLElement) {
	// log.Println("authenticate.OnHTML()")
	// log.Println(e.Attr("content"))
	// })

	router.collector.OnResponse(func(r *colly.Response) {
		// log.Println("authenticate.OnResponse()")
		// print cookie
		// fmt.Println(router.collector.Cookies(r.Request.URL.String()))
		if r.StatusCode != 200 {
			err = fmt.Errorf("status code error: %d %s", r.StatusCode, r.Body)
			return
		}

		// var jsonResponse map[string]string
		err = json.Unmarshal([]byte(r.Body), &response)
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("authenticate.OnError()")
		log.Println("authenticate => error:", err, r.Body)
	})

	err = router.collector.Post(url, data)

	// router.collector.OnHTMLDetach("head > meta:nth-child(3)")
	return
}

func (router *TechnicolorRouter) getSAndB(authenticateResponse map[string]string) (err error) {
	// type SAndBResponse struct {
	// 	S string `json:"s"`
	// 	B string `json:"b"`
	// }

	if _, ok := authenticateResponse["s"]; !ok {
		err = fmt.Errorf("s not found in response")
		return
	}

	if _, ok := authenticateResponse["B"]; !ok {
		err = fmt.Errorf("B not found in response")
		return
	}

	router.s, err = hex.DecodeString(authenticateResponse["s"])
	router.B, err = hex.DecodeString(authenticateResponse["B"])
	return
}

func (router *TechnicolorRouter) getM(authenticateResponse map[string]string) (err error) {
	_, containsError := authenticateResponse["error"]

	if containsError {
		err = fmt.Errorf("error in response: %v", authenticateResponse)
		return
	}

	if _, ok := authenticateResponse["M"]; !ok {
		err = fmt.Errorf("M not found in response")
		return
	}
	router.M, err = hex.DecodeString(authenticateResponse["M"])
	return
}
