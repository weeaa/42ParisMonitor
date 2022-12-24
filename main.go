package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
	
	wh "github.com/etaaa/go-webhooks"
)

// 🚨 If you need help using the monitor, you can reach out to me on Discord at weeaa#4144 (363975551393988620) 🚨

// Email42, Password42, etc... must be modified!
const Email42 = "xavierniel@free.fr"                                                  // ☎️
const Password42 = "d9babcc8b57b6a7455b5220fe8346de7e2db06df29ba58c855ab5241e49b3bbb" // ;)
const DiscordID = "363975551393988620"                                                //weeaa#4144 on Discord
const DiscordWebhook = "https://discord.com/api/webhooks/..."

// OperatingSystem and UserAgent let you customize headers (you can leave it as it is)
const (
	OperatingSystem = "\"macOS\""
	UserAgent       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
)

type User struct {
	Email        string
	Password     string
	DefaultSleep int
	Cookies      Cookies
	Settings     RequestSettings
}

type RequestSettings struct {
	OS string
	UA string
}

type Cookies struct {
	CSRF                        string
	AdmissionsSessionProduction string
}

// keywords will be used to detect if a Piscine 🏊 is open
var keywords = []string{
	"July",
	"June",
	"August",
	"March",
	"April",
}

var (
	jar, _ = cookiejar.New(nil)
	client = http.Client{
		Jar: jar,
	}
)

var (
	invalidCredentials         = errors.New("ERR Invalid Credentials/Unknown")
	unableGetCookies           = errors.New("ERR Cookies Are not in the Response")
	unableGetAuthenticityToken = errors.New("ERR Getting Authenticity Token")
)

func main() {
	piscineFound := false
	needLogin := true
	
	task := User{
		Email:        Email42,
		Password:     Password42,
		DefaultSleep: 5,
		Settings: RequestSettings{
			UA: UserAgent,
			OS: OperatingSystem,
		},
	}
	
	uri, _ := url.Parse("https://admissions.42.fr/campus/paris/campus_steps/42-paris-piscine-8")
	
	log.Println("Launched 42Paris Monitor!")
	
	for {
		
		if needLogin {
			err := task.login42Paris()
			if err != nil {
				task.defaultSleep(err)
				continue
			}
		}
		
		needLogin = false
		
		req := http.Request{
			Method: http.MethodGet,
			URL:    uri,
			Header: http.Header{
				"authority":                 {"admissions.42.fr"},
				"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,/;q=0.8,application/signed-exchange;v=b3;q=0.9"},
				"accept-language":           {"fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7"},
				"cache-control":             {"max-age=0"},
				"content-type":              {"application/x-www-form-urlencoded"},
				"cookie":                    {"locale=fr; _admissions_session_production=" + task.Cookies.AdmissionsSessionProduction},
				"referer":                   {"https://admissions.42.fr/users/sign_in"},
				"sec-ch-ua":                 {"\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Google Chrome\";v=\"108"},
				"sec-ch-ua-mobile":          {"?0"},
				"sec-ch-ua-platform":        {task.Settings.OS},
				"sec-fetch-dest":            {"document"},
				"sec-fetch-mode":            {"navigate"},
				"sec-fetch-site":            {"same-origin"},
				"sec-fetch-user":            {"?1"},
				"upgrade-insecure-requests": {"1"},
				"user-agent":                {task.Settings.UA},
			},
		}
		
		resp, err := client.Do(&req)
		if err != nil {
			task.defaultSleep(err)
		}
		
		body, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode != 200 {
			task.defaultSleep(fmt.Sprintf("ERR Unknown Fetching Piscine Availability [%v]", resp.Status))
			continue
		} else if strings.Contains(string(body), "42 Paris | Connexion") {
			needLogin = true
			continue
		}
		
		// First Check
		for _, month := range keywords {
			if strings.Contains(string(body), month) {
				log.Printf("A Piscine is Open in %v!", month)
				go task.sendDiscordNotification()
				piscineFound = true
			}
		}
		
		// will be done soon
		/*
			page, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
			if err != nil {
		
			}
			page.Find("li.list-group-item")
		
		*/
		
		if piscineFound {
			break
		}
		
		resp.Body.Close()
		
		// does not have a piscine open, so we retry
		time.Sleep(2400 * time.Millisecond)
		
	}
	
	task.safeExit("Process Ending...")
}

func (t *User) login42Paris() (err error) {
	
	//get csrf token
	uri, _ := url.Parse("https://admissions.42.fr/")
	req := http.Request{
		Method: http.MethodGet,
		URL:    uri,
		Host:   "admissions.42.fr",
		Header: http.Header{
			"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,/;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			"accept-language":           {"fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7"},
			"accept-encoding":           {"gzip, deflate, br"},
			"connection":                {"keep-alive"},
			"referer":                   {"https://42.fr/en/homepage/"},
			"sec-ch-ua":                 {"\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Google Chrome\";v=\"108\""},
			"sec-ch-ua-mobile":          {"?0"},
			"sec-ch-ua-platform":        {t.Settings.OS},
			"sec-fetch-dest":            {"document"},
			"sec-fetch-mode":            {"navigate"},
			"sec-fetch-site":            {"same-origin"},
			"sec-fetch-user":            {"?1"},
			"upgrade-insecure-requests": {"1"},
			"user-agent":                {t.Settings.UA},
		},
	}
	
	resp, err := client.Do(&req)
	if err != nil {
		return err
	}
	
	log.Println(resp.Cookies())
	
	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == "_admissions_session_production" {
			t.Cookies.AdmissionsSessionProduction = c.Value
			log.Println("Got AdmissionsSessionProduction Cookie")
		}
		
	}
	
	resp.Body.Close()
	
	//get authenticity token
	uri, _ = url.Parse("https://admissions.42.fr/users/sign_in")
	
	req = http.Request{
		Method: http.MethodGet,
		URL:    uri,
		Header: http.Header{
			"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,/;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			"accept-language":           {"fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7"},
			"connection":                {"keep-alive"},
			"cookie":                    {"locale=fr; _admissions_session_production=" + t.Cookies.AdmissionsSessionProduction},
			"referer":                   {"https://42.fr/en/homepage/"},
			"sec-ch-ua":                 {"\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Google Chrome\";v=\"108"},
			"sec-ch-ua-mobile":          {"?0"},
			"sec-ch-ua-platform":        {t.Settings.OS},
			"sec-fetch-dest":            {"document"},
			"sec-fetch-mode":            {"navigate"},
			"sec-fetch-site":            {"same-site"},
			"sec-fetch-user":            {"?1"},
			"upgrade-insecure-requests": {"1"},
			"user-agent":                {t.Settings.UA},
		},
	}
	
	resp, err = client.Do(&req)
	if err != nil {
		return err
	}
	
	body, _ := io.ReadAll(resp.Body)
	
	page, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	
	t.Cookies.CSRF = page.Find("input[name=authenticity_token]").AttrOr("value", "")
	if t.Cookies.CSRF == "" {
		return unableGetAuthenticityToken
	} else {
		log.Println("Got Authenticity Token")
	}
	
	resp.Body.Close()
	
	//sign in
	uri, _ = url.Parse("https://admissions.42.fr/users/sign_in")
	
	params := url.Values{
		"utf8":               {"✓"},
		"authenticity_token": {t.Cookies.CSRF},
		"user[email]":        {t.Email},
		"user[password]":     {t.Password},
	}
	
	payload := strings.NewReader(params.Encode())
	
	req = http.Request{
		Method: http.MethodPost,
		URL:    uri,
		Body:   io.NopCloser(payload),
		Header: http.Header{
			"authority":                 {"admissions.42.fr"},
			"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,/;q=0.8,application/signed-exchange;v=b3;q=0.9"},
			"accept-language":           {"fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7"},
			"cache-control":             {"max-age=0"},
			"content-type":              {"application/x-www-form-urlencoded"},
			"cookie":                    {"locale=fr; _admissions_session_production=" + t.Cookies.AdmissionsSessionProduction},
			"origin":                    {"https://admissions.42.fr/"},
			"referer":                   {"https://admissions.42.fr/users/sign_in"},
			"sec-ch-ua":                 {"\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Google Chrome\";v=\"108"},
			"sec-ch-ua-mobile":          {"?0"},
			"sec-ch-ua-platform":        {t.Settings.OS},
			"sec-fetch-dest":            {"document"},
			"sec-fetch-mode":            {"navigate"},
			"sec-fetch-site":            {"same-origin"},
			"sec-fetch-user":            {"?1"},
			"upgrade-insecure-requests": {"1"},
			"user-agent":                {t.Settings.UA},
		},
	}
	
	resp, err = client.Do(&req)
	if err != nil {
		return err
	}
	
	if resp.StatusCode != 302 {
		return invalidCredentials
	}
	
	resp.Body.Close()
	
	log.Println("Logged In!")
	
	return nil
}

func (t *User) sendDiscordNotification() {
	
	piscineURL := "https://admissions.42.fr/campus/paris/campus_steps/42-paris-piscine-8"
	loginURL := "[Login](https://admissions.42.fr/users/sign_in)"
	
	webhook := wh.Webhook{
		Content:   "<@" + DiscordID + ">",
		Username:  "42Paris Monitor",
		AvatarUrl: "https://42.fr/wp-content/uploads/2021/08/42.jpg",
		Embeds: []wh.Embed{
			{
				Title:     "A Piscine is Available!",
				Url:       piscineURL,
				Timestamp: wh.GetTimestamp(),
				Color:     wh.GetColor("#7fe6eb"),
				Footer: wh.EmbedFooter{
					Text: "42Paris Monitor | by weeaa#4144",
				},
				Thumbnail: wh.EmbedThumbnail{
					Url: "https://emojipedia-us.s3.amazonaws.com/source/microsoft-teams/337/person-swimming_1f3ca.png",
				},
				Author: wh.EmbedAuthor{
					Name:    "@weea_a",
					Url:     "https://twitter.com/weea_a",
					IconUrl: "https://pbs.twimg.com/profile_images/1591885204137951233/7xfFjgH4_400x400.jpg",
				},
				Fields: []wh.EmbedFields{
					{
						Name:  "QuickLinks",
						Value: loginURL,
					},
				},
			},
		},
	}
	if err := wh.SendWebhook(DiscordWebhook, webhook, false); err != nil {
		log.Println(err)
	}
}

func (t *User) safeExit(msg any) {
	fmt.Println(msg)
	time.Sleep(time.Duration(t.DefaultSleep) * time.Second)
	os.Exit(0)
}

func (t *User) defaultSleep(msg any) {
	log.Println(msg)
	time.Sleep(time.Duration(t.DefaultSleep) * time.Second)
}
