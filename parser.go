package main

import (
	"fmt"
	"log"
	"os"
	"time"

	s "github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var pages int = 1
var pageIndex int = 1
var data []string
var games []Game
type Game struct{
	Name string `json:"name"`
	Url string `json:"url"`
	Avg string `avg`
}


func main() {
	service, err := s.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		log.Fatal("Не удалось запустить ChromeDriver:", err)
	}
	defer service.Stop()

	caps := s.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"window-size=1920x1080",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-blink-features=AutomationControlled",
			"--disable-infobars",
		},
	})

	driver, err := s.NewRemote(caps, "")
	if err != nil {
		log.Fatal("Не удалось подключиться к WebDriver:", err)
	}
	defer driver.Quit()

	err = driver.Get("https://www.backloggd.com/games/lib/popular?page=1")
	if err != nil {
		log.Fatal("Не удалось загрузить страницу:", err)
	}

	for{
		collectDataFromPage(driver)
		if pageIndex != pages{
			
			pageIndex += 1
			pageStr := fmt.Sprintf("https://www.backloggd.com/games/lib/popular?page=%d", pageIndex)
			log.Println(pageStr)
			err = driver.Get(pageStr)
			if err != nil {
				log.Fatal("Не удалось загрузить страницу:", err)
			}
		}else{
			break
		}
	}
	addDataInFile()
	time.Sleep(1 * time.Second)
}

func collectDataFromPage(driver s.WebDriver){


	for i := 0; i < 2; i++ {
		_, err := driver.ExecuteScript("window.scrollBy(0, 500);", nil)
		if err != nil {
			log.Println("Ошибка при скроллинге:", err)
			return
		}
		time.Sleep(200 * time.Millisecond)
	}

	main, err := driver.FindElement(s.ByCSSSelector, "main.main")
	if err != nil {
		log.Println("Не удалось найти container:", err)
		return
	}

	container, err := main.FindElement(s.ByCSSSelector, "div.container")
	if err != nil {
		log.Println("Не удалось найти container:", err)
		return
	}

	games, err := container.FindElements(s.ByXPATH, "./div[4]")
	if err != nil {
		log.Println("Не удалось найти div по позиции:", err)
		return
	}

	for _, elem := range games {
		links, err := elem.FindElements(s.ByCSSSelector, "a")
		if err != nil {
			log.Println("Не удалось найти ссылки в контейнере:", err)
			continue
		}

		var lastHref string
		for _, link := range links {
			href, err := link.GetAttribute("href")
			if err != nil {
				log.Println("Не удалось получить href:", err)
				continue
			}

			if href != lastHref {
				takeDataFromGame(href, driver)
				lastHref = href
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func takeDataFromGame(href string, driver s.WebDriver) {
	err := driver.Get(href)
	if err != nil {
		log.Println("не удалось перейти", err)
		return
	}

	main, err := driver.FindElement(s.ByCSSSelector, "main.main")
	if err != nil {
		log.Println("Не удалось найти container:", err)
		return
	}

	container, err := main.FindElement(s.ByCSSSelector, "div.container")
	if err != nil {
		log.Println("Не удалось найти container:", err)
		return
	}

	info, err := container.FindElements(s.ByXPATH, "./div[4]")
	if err != nil {
		log.Println("Не удалось найти div по позиции:", err)
		return
	}

	for _, elemGame := range info {
		h1, err := elemGame.FindElements(s.ByCSSSelector, "h1")
		if err != nil {
			log.Println("Не удалось найти avg:", err)
			return
		}

		avg, _ := h1[1].Text()
		name, _ := h1[3].Text()
		line := fmt.Sprintf("Название игры: %s, Ссылка: %s , avg: %s", name, href, avg)
		log.Println(line)
		var game Game
		game.Name = name
		game.Url = href
		game.Avg = avg
		games = append(games, game)
	}

	err = driver.Back()
	if err != nil {
		log.Println("Не удалось вернуться на исходный сайт:", err)
		return
	}
}

func addDataInFile() {
	file, err := os.OpenFile("games.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Не удалось открыть файл:", err)
		return
	}
	defer file.Close()

	for _, game := range games {
		line := fmt.Sprintf("{\"name\": \"%s\", \"url\": \"%s\", \"avg\": \"%s\"}\n", game.Name, game.Url, game.Avg)
		_, err := file.WriteString(line)
		if err != nil {
			log.Println("Ошибка при записи в файл:", err)
			return
		}
	}
	log.Println("Данные успешно добавлены в файл games.json")
}