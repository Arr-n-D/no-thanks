package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var (
	linux = flag.Bool("linux", false, "If true, use Linux drivers and commands.")
)

type Account struct {
	Username string
	Password string
}

func main() {
	Start()
}

func Initialize() selenium.WebDriver {
	const (
		// These paths will be different on your system.
		seleniumPath      = "browsers/selenium-server.jar"
		chromeDriver      = "browsers/chromedriver.exe"
		linuxChromeDriver = "browsers/chromedriver"
		port              = 8080
	)
	opts := []selenium.ServiceOption{
		// selenium.StartFrameBuffer(),         // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriver), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),          // Output debug information to STDERR.
	}
	// selenium.SetDebug(true)

	if *linux {
		opts = append(opts, selenium.ChromeDriver(linuxChromeDriver))
		opts = append(opts, selenium.StartFrameBuffer())
	}

	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{"browserName": "chrome"}

	if *linux {
		chromeCaps := chrome.Capabilities{
			Path: "",
			Args: []string{
				"--start-maximized",
				"--window-size=1200x600",
				"--no-sandbox",
				"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			},
		}
		caps.AddChrome(chromeCaps)
	}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}

	return wd
}

func Start() {
	// Start a Selenium WebDriver server instance (if one is not already
	// running).

	wd := Initialize()
	defer wd.Quit()

	GoToLinkedIn(wd)

	// Navigate to the simple playground interface.
	// if err := wd.Get("http://play.golang.org/?simple=1"); err != nil {
	// 	panic(err)
	// }

	// // Get a reference to the text box containing code.
	// elem, err := wd.FindElement(selenium.ByCSSSelector, "#code")
	// if err != nil {
	// 	panic(err)
	// }
	// // Remove the boilerplate code already in the text box.
	// if err := elem.Clear(); err != nil {
	// 	panic(err)
	// }

	// // Enter some new code in text box.
	// err = elem.SendKeys(`
	// 	package main
	// 	import "fmt"
	// 	func main() {
	// 		fmt.Println("Hello WebDriver!")
	// 	}
	// `)
	// if err != nil {
	// 	panic(err)
	// }

	// // Click the run button.
	// btn, err := wd.FindElement(selenium.ByCSSSelector, "#run")
	// if err != nil {
	// 	panic(err)
	// }
	// if err := btn.Click(); err != nil {
	// 	panic(err)
	// }

	// // Wait for the program to finish running and get the output.
	// outputDiv, err := wd.FindElement(selenium.ByCSSSelector, "#output")
	// if err != nil {
	// 	panic(err)
	// }

	// var output string
	// for {
	// 	output, err = outputDiv.Text()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if output != "Waiting for remote server..." {
	// 		break
	// 	}
	// 	time.Sleep(time.Millisecond * 100)
	// }

	// fmt.Printf("%s", strings.Replace(output, "\n\n", "\n", -1))
	// // Example Output:
	// // Hello WebDriver!
	// //
	// // Program exited.

	// // The following shows an example of using the Actions API.
	// // Please refer to the WC3 Actions spec for more detailed information.
	// if err := wd.Get("http://play.golang.org/?simple=1"); err != nil {
	// 	panic(err)
	// }

}

func GoToLinkedIn(wd selenium.WebDriver) {
	// Navigate to the simple playground interface.
	if err := wd.Get("https://www.linkedin.com/"); err != nil {
		panic(err)
	}

	LoginToLinkedIn(wd)

	time.Sleep(time.Second * 30)
}

func LoginToLinkedIn(wd selenium.WebDriver) {

	// find the element that's ID attribute is 'session_key'
	elem, err := wd.FindElement(selenium.ByID, "session_key")
	if err != nil {
		panic(err)
	}

	// read the username from the credentials.json file
	file, _ := ioutil.ReadFile("credentials.json")

	credentials := Account{}

	_ = json.Unmarshal([]byte(file), &credentials)

	// enter the username in the field
	err = elem.SendKeys(credentials.Username)
	if err != nil {
		panic(err)
	}

	// find the element that's ID attribute is 'session_password'
	elem, err = wd.FindElement(selenium.ByID, "session_password")
	if err != nil {
		panic(err)
	}

	// enter the password in the field
	err = elem.SendKeys(credentials.Password)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 5)

	// find the sign in button by its class attribute
	elem, err = wd.FindElement(selenium.ByCSSSelector, ".sign-in-form__submit-button")
	if err != nil {
		panic(err)
	}

	// click the button
	err = elem.Click()

	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 1)
	//wait to find msg-overlay-list-bubble-search
	elem, err = wd.FindElement(selenium.ByCSSSelector, ".msg-conversations-container__conversations-list.msg-overlay-list-bubble__conversations-list")
	if err != nil {
		panic(err)
	}

	// find the scrollable section with class msg-overlay-list-bubble__content--scrollable
	section, err := wd.FindElement(selenium.ByCSSSelector, ".msg-overlay-list-bubble__content--scrollable")
	if err != nil {
		panic(err)
	}

	// loop 20 times and scroll down the page
	// look, there's a better way to do this, like checking positions, BUT, this is a quick and dirty way to do it and that works for me
	for i := 0; i < 20; i++ {
		ScrollToBottomOfMessages(wd, section)
		time.Sleep(time.Millisecond * 500)
	}

	messages, err := elem.FindElements(selenium.ByCSSSelector, ".msg-conversation-listitem__link")
	if err != nil {
		panic(err)
	}

	// for each messages in the list, loop through and check if the message is a new message msg-overlay-list-bubble__message-snippet--v2 m0 t-black t-12 t-bold
	for _, message := range messages {
		elem, err := message.FindElement(selenium.ByCSSSelector, ".msg-overlay-list-bubble__message-snippet--v2.m0.t-black.t-12.t-bold")
		if err != nil {
			fmt.Println("Didn't find new message for this element")
			continue
		}

		// if the message is a new message, click it
		elem.Click()
	}

}

func ScrollToBottomOfMessages(wd selenium.WebDriver, section selenium.WebElement) {
	args := []interface{}{section}
	// section
	// scroll down to bottom of section
	// javascript executor to scroll to bottom of section
	js := "arguments[0].scrollTop = arguments[0].scrollHeight"
	_, err := wd.ExecuteScript(js, args)

	if err != nil {
		panic(err)
	}
}
