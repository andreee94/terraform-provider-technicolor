package technicolor

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

func Bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (router *TechnicolorRouter) DeletePortForwarded(index int) (err error) {
	url := router.getEndpoint(TECHNICOLOR_ENDPOINT_PORT_FORWARDING)
	// url := fmt.Sprintf("%s/modals/wanservices-modal.lp", router.url)

	router.collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			err = fmt.Errorf("status code error: %d %s", r.StatusCode, r.Body)
			return
		}
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("DeletePortForwarding.OnError()")
		log.Println("DeletePortForwarding => error:", err, r.Body)
	})

	data := map[string]string{
		"tableid":   "portforwarding",
		"stateid":   "",
		"action":    "TABLE-DELETE",
		"index":     fmt.Sprintf("%d", index),
		"CSRFtoken": router.CSRFToken,
	}

	err = router.collector.Post(url, data)

	router.collector.OnHTMLDetach("head > meta:nth-child(3)")
	return
}

func (router *TechnicolorRouter) AddPortForwarded(newPortForwarding *PortForwarded) (err error) {
	url := router.getEndpoint(TECHNICOLOR_ENDPOINT_PORT_FORWARDING)
	// url := fmt.Sprintf("%s/modals/wanservices-modal.lp", router.url)

	router.collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			err = fmt.Errorf("status code error: %d %s", r.StatusCode, r.Body)
			return
		}
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("AddPortForwarding.OnError()")
		log.Println("AddPortForwarding => error:", err, r.Body)
	})

	err, portsForwarded := router.GetAllPortForwarded()

	index := len(portsForwarded) + 1 // index starts from 1

	data := map[string]string{
		"enabled":       fmt.Sprintf("%d", Bool2int(newPortForwarding.Enabled)),
		"name":          newPortForwarding.Name,
		"protocol":      newPortForwarding.Protocol,
		"wanport":       fmt.Sprintf("%d", newPortForwarding.WanPort),
		"lanport":       fmt.Sprintf("%d", newPortForwarding.LanPort),
		"destinationip": newPortForwarding.LanIp,
		"tableid":       "portforwarding",
		"stateid":       "",
		"action":        "TABLE-ADD",
		"index":         fmt.Sprintf("%d", index),
		"CSRFtoken":     router.CSRFToken,
	}

	err = router.collector.Post(url, data)
	return
}

func (router *TechnicolorRouter) GetAllPortForwarded() (err error, portsForwarded []PortForwardedWithIndex) {
	url := router.getEndpoint(TECHNICOLOR_ENDPOINT_PORT_FORWARDING)
	// url := fmt.Sprintf("%s/modals/wanservices-modal.lp", router.url)

	router.collector.OnHTML("#portforwarding > tbody:nth-child(2)", func(e *colly.HTMLElement) {
		// log.Println("GetAllPortForwarded.OnHTML()")
		var index = 1 // index starts from 1
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			portForwarded := PortForwarded{}
			el.ForEach("td", func(j int, el2 *colly.HTMLElement) {
				switch j {
				case 0:
					portForwarded.Enabled = el2.ChildAttr("input", "value") == "1"
				case 1:
					portForwarded.Name = el2.Text
				case 2:
					portForwarded.Protocol = el2.Text
				case 3:
					portForwarded.WanPort, err = parsePort(el2.Text)
				case 4:
					portForwarded.LanPort, err = parsePort(el2.Text)
				// case 5: ???
				case 6:
					portForwarded.LanIp = strings.TrimSpace(el2.Text)
				case 7:
					portForwarded.LanMac = strings.TrimSpace(el2.Text)
				}
			})
			// log.Printf("%+v", portForwarded)
			portsForwarded = append(portsForwarded, PortForwardedWithIndex{
				Index: index,
				Data:  portForwarded,
			})
			index++
		})
	})

	router.collector.OnResponse(func(r *colly.Response) {
		if r.StatusCode != 200 {
			err = fmt.Errorf("status code error: %d %s", r.StatusCode, r.Body)
			return
		}
	})

	router.collector.OnError(func(r *colly.Response, err error) {
		log.Println("GetAllPortForwarded.OnError()")
		log.Println("GetAllPortForwarded => error:", err, r.Body)
	})

	router.collector.Visit(url)

	router.collector.OnHTMLDetach("#portforwarding > tbody:nth-child(2)")
	return
}

func (router *TechnicolorRouter) GetPortForwardedByName(name string) (err error, portForwarded PortForwardedWithIndex) {
	err, portsForwarded := router.GetAllPortForwarded()

	for _, portForwarded := range portsForwarded {
		if portForwarded.Data.Name == name {
			return nil, portForwarded
			// return nil, PortForwardedWithIndex{
			// 	Index: index + 1, // index starts from 1
			// 	Data:  portForwarded.Data,
			// }
		}
	}

	return fmt.Errorf("port forwarding not found"), PortForwardedWithIndex{Index: -1}
}

var PORT_REGEX = regexp.MustCompile(`.*\((?P<port>\d+)\)`)

func parsePort(portString string) (port int, err error) {
	port, err = strconv.Atoi(portString)

	if err == nil {
		return port, nil
	}

	if len(PORT_REGEX.FindStringIndex(portString)) > 0 {
		var portStringExtracted = PORT_REGEX.FindStringSubmatch(portString)[1]

		port, err = strconv.Atoi(portStringExtracted)
	}
	return
}
