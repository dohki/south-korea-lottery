package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/tebeka/selenium"
)

// Import from config
const (
	driverPath = "/Users/dohki/Desktop/chromedriver"
	port       = 8080
	host       = "https://www.dhlottery.co.kr"
)

var hostURL *url.URL
var timeout time.Duration

func init() {
	var err error
	if hostURL, err = url.Parse(host); err != nil {
		panic(err)
	}

	timeout = 3 * time.Second
}

func debug(arg interface{}) {
	fmt.Println(arg)
	fmt.Scanln()
}

func panicAtError(err error) {
	if err != nil {
		panic(err)
	}
}

func initWebDriver() (*selenium.Service, selenium.WebDriver) {
	// TODO: Offer various browsers
	// TODO: Make verbose option in tebeka/selenium
	opts := []selenium.ServiceOption{
		selenium.ChromeDriver(driverPath),
		selenium.Output(os.Stderr),
	}

	service, err := selenium.NewChromeDriverService(driverPath, port, opts...)
	panicAtError(err)

	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	panicAtError(err)

	return service, wd
}

func isWebPageLoaded(wd selenium.WebDriver) (bool, error) {
	result, err := wd.ExecuteScript("return document.readyState", nil)
	return result == "complete", err
}

func login(wd selenium.WebDriver) {
	err := wd.Get(host)
	panicAtError(err)
	err = wd.WaitWithTimeout(isWebPageLoaded, timeout)
	panicAtError(err)

	elem, err := wd.FindElement(selenium.ByCSSSelector, "a[href*=login]")
	panicAtError(err)
	err = elem.Click()
	panicAtError(err)

	elem, err = wd.FindElement(selenium.ByCSSSelector, "input[name=userId]")
	panicAtError(err)
	// TODO: Import from config
	err = elem.SendKeys(os.Getenv("SKL_USER_ID"))
	panicAtError(err)

	elem, err = wd.FindElement(selenium.ByCSSSelector, "input[name=password]")
	panicAtError(err)
	err = elem.SendKeys(os.Getenv("SKL_USER_PW"))
	panicAtError(err)

	elem, err = wd.FindElement(selenium.ByCSSSelector, "a[href*=check_if_Valid3]")
	panicAtError(err)
	err = elem.Click()
	panicAtError(err)
}

func buyLotto645(wd selenium.WebDriver) {
	/*
		err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
			_, err := wd.FindElement(selenium.ByCSSSelector, "a[href*=buyLotto]")
			return err != nil, err
		}, timeout)
	*/
	err := wd.WaitWithTimeout(isWebPageLoaded, timeout)
	panicAtError(err)

	// FIXME: Find robust way to synchronize
	elem, err := wd.FindElement(selenium.ByCSSSelector, "a[href*=buyLotto]")
	panicAtError(err)
	// WebElement.Click() dies with "element not interactable"
	href, err := elem.GetAttribute("href")
	panicAtError(err)
	err = wd.Get(href)
	panicAtError(err)
	err = wd.WaitWithTimeout(isWebPageLoaded, timeout)
	panicAtError(err)

	_, err = wd.ExecuteScript("goLottoBuy(1)", nil)
	panicAtError(err)
	err = wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		winHdls, err := wd.WindowHandles()
		if err != nil {
			return false, err
		}
		if len(winHdls) == 1 {
			return false, nil
		}
		err = wd.SwitchWindow(winHdls[1])
		return true, err
	}, timeout)
	panicAtError(err)

	elem, err = wd.FindElement(selenium.ByCSSSelector, "#checkAutoSelect")
	panicAtError(err)
	err = elem.Click()
	panicAtError(err)

	// TODO: Buy
}

// TODO: Split into multiple packages
func main() {
	service, wd := initWebDriver()
	defer service.Stop()
	defer wd.Quit()

	login(wd)
	buyLotto645(wd)

	fmt.Scanln()
}
