package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/google/martian/log"
	"github.com/google/martian/v3/log"
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
		// selenium.Output(os.Stderr),          // Output debug information to STDERR.
	}
	selenium.SetDebug(false)

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
	LoginToLinkedIn(wd)

	// infinite go routine
	// go func([]selenium.WebElement, selenium.WebDriver) {
	for {
	LABEL:
		messages := fetchMessages(wd)
		messages = loopMessages(messages, wd)

		if len(messages) == 0 {
			goto LABEL
		}
		log.Error

	}
	// }(messages, wd)

}

func GoToLinkedIn(wd selenium.WebDriver) {
	// Navigate to the simple playground interface.
	if err := wd.Get("https://www.linkedin.com/"); err != nil {
		panic(err)
	}

	time.Sleep(time.Millisecond * 500)

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

}

func loopMessages(messages []selenium.WebElement, wd selenium.WebDriver) []selenium.WebElement {
	for _, message := range messages {
		elem, err := message.FindElement(selenium.ByCSSSelector, ".msg-overlay-list-bubble__message-snippet--v2.m0.t-black.t-12.t-bold")
		if err != nil {
			// remove the message from the messages slice
			// print the index and index + 1
			messages = messages[1:]
			continue
		} else {
			handleMessage(elem, wd)
			messages = messages[1:]
		}

	}

	return messages
}

func handleMessage(elem selenium.WebElement, wd selenium.WebDriver) {
	elem.Click()

	conversationBubble, err := wd.FindElement(selenium.ByCSSSelector, ".msg-overlay-conversation-bubble--is-active")
	if err != nil {
		panic(err)
	}

	elem, err = wd.FindElement(selenium.ByCSSSelector, ".msg-s-message-list__event.clearfix.msg-s-message-list__event--slide-in")
	if err != nil {
		panic(err)
	}

	messageData, err := elem.FindElement(selenium.ByCSSSelector, ".msg-s-event-listitem__body.t-14.t-black--light.t-normal")
	if err != nil {
		panic(err)
	}

	messageText, err := messageData.Text()
	fmt.Println(messageText)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 2)

	replyToMessageIfMatchesKeyword(messageText, conversationBubble)

}

func replyToMessageIfMatchesKeyword(messageText string, conversationBubble selenium.WebElement) {
	fmt.Println("Checking if message matches keyword")
	file, err := os.Open("keywords/keywords.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var keywords []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keywords = append(keywords, scanner.Text())
	}

	for _, keyword := range keywords {
		fmt.Println(keyword)
		replyToMessage(messageText, keyword, conversationBubble)
	}
}

func replyToMessage(messageText string, keyword string, conversationBubble selenium.WebElement) {
	fmt.Println("Looking to reply to: " + messageText + " by matching keyword: " + keyword)
	fmt.Println(messageText, keyword)
	if strings.Contains(messageText, keyword) {
		fmt.Println("Found keyword: " + keyword)

		file, _ := ioutil.ReadFile("automated_replies/reply.txt")
		reply := string(file)

		editableInput, err := conversationBubble.FindElement(selenium.ByCSSSelector, ".msg-form__contenteditable.t-14.t-black--light.t-normal.flex-grow-1.full-height.notranslate")
		if err != nil {
			panic(err)
		}

		err = editableInput.SendKeys(reply)
		if err != nil {
			panic(err)
		}

		sendButton, err := conversationBubble.FindElement(selenium.ByCSSSelector, ".msg-form__send-button.artdeco-button.artdeco-button--1")
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Millisecond * 300)
		err = sendButton.Click()
		if err != nil {
			panic(err)
		}

	} else {
		fmt.Println("No keyword found")
	}
}

func fetchMessages(wd selenium.WebDriver) []selenium.WebElement {
	fmt.Println("Fetching messages...")
	elem, err := wd.FindElement(selenium.ByCSSSelector, ".msg-conversations-container__conversations-list.msg-overlay-list-bubble__conversations-list")
	if err != nil {
		panic(err)
	}

	section, err := wd.FindElement(selenium.ByCSSSelector, ".msg-overlay-list-bubble__content--scrollable")
	if err != nil {
		panic(err)
	}

	// loop 20 times and scroll down the page
	// look, there's a better way to do this, like checking positions, BUT, this is a quick and dirty way to do it and that works for me
	for i := 0; i < 20; i++ {
		scrollToBottomOfMessages(wd, section)
		time.Sleep(time.Millisecond * 500)
	}

	messages, err := elem.FindElements(selenium.ByCSSSelector, ".msg-conversation-listitem__link")
	if err != nil {
		panic(err)
	}
	return messages
}

func scrollToBottomOfMessages(wd selenium.WebDriver, section selenium.WebElement) {
	args := []interface{}{section}
	js := "arguments[0].scrollTop = arguments[0].scrollHeight"
	_, err := wd.ExecuteScript(js, args)

	if err != nil {
		panic(err)
	}
}
